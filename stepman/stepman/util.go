package stepman

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

// DebugMode ...
var DebugMode bool

func parseStepYml(collectionURI, pth, id, version string) (models.StepModel, error) {
	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.StepModel{}, err
	}

	var stepModel models.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		return models.StepModel{}, err
	}

	if err := stepModel.Normalize(); err != nil {
		return models.StepModel{}, err
	}

	if err := stepModel.ValidateStep(); err != nil {
		return models.StepModel{}, err
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return models.StepModel{}, err
	}

	return stepModel, nil
}

// ParseStepCollection ...
func ParseStepCollection(pth string) (models.StepCollectionModel, error) {
	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.StepCollectionModel{}, err
	}

	var stepCollection models.StepCollectionModel
	if err := yaml.Unmarshal(bytes, &stepCollection); err != nil {
		return models.StepCollectionModel{}, err
	}
	return stepCollection, nil
}

// DownloadStep ...
func DownloadStep(collection models.StepCollectionModel, id, version, commithash string) error {
	log.Debugf("Download Step: %#v (%#v)\n", id, version)
	downloadLocations, err := collection.GetDownloadLocations(id, version)
	if err != nil {
		return err
	}

	route, found := ReadRoute(collection.SteplibSource)
	if !found {
		return errors.New("No routing found for lib: " + err.Error())
	}

	stepPth := GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepPth); err != nil {
		return err
	} else if exist {
		log.Debug("[STEPMAN] - Step already downloaded")
		return nil
	}

	success := false
	for _, downloadLocation := range downloadLocations {
		switch downloadLocation.Type {
		case "zip":
			log.Debug("[STEPMAN] - Downloading step from:", downloadLocation.Src)
			if err := DownloadAndUnZIP(downloadLocation.Src, stepPth); err != nil {
				log.Warn("[STEPMAN] - Failed to download step.zip:", err)
			} else {
				success = true
				return nil
			}
		case "git":
			log.Debug("[STEPMAN] - Git clone step from:", downloadLocation.Src)
			if err := DoGitCloneWithCommit(downloadLocation.Src, stepPth, version, commithash); err != nil {
				log.Warn("[STEPMAN] - Failed to clone step (%s): %v", downloadLocation.Src, err)
			} else {
				success = true
				return nil
			}
		default:
			return fmt.Errorf("[STEPMAN] - Failed to download: Invalid download location (%#v) for step %#v (%#v)", downloadLocation, id, version)
		}
	}

	if !success {
		return errors.New("Failed to download step")
	}
	return nil
}

func addStepVersionToStepGroup(step models.StepModel, version string, stepGroup models.StepGroupModel) (models.StepGroupModel, error) {
	if stepGroup.LatestVersionNumber != "" {
		r, err := models.CompareVersions(stepGroup.LatestVersionNumber, version)
		if err != nil {
			return models.StepGroupModel{}, err
		}
		if r == 1 {
			stepGroup.LatestVersionNumber = version
		}
	} else {
		stepGroup.LatestVersionNumber = version
	}
	log.Debugf("SetGroup: %#v, versionParam: %#v, stepParam: %#v", stepGroup, version, step)
	stepGroup.Versions[version] = step
	return stepGroup, nil
}

func generateFormattedJSONForStepsSpec(route SteplibRoute, templateCollection models.StepCollectionModel) ([]byte, error) {
	collection := models.StepCollectionModel{
		FormatVersion:        templateCollection.FormatVersion,
		GeneratedAtTimeStamp: time.Now().Unix(),
		SteplibSource:        route.SteplibURI,
		DownloadLocations:    templateCollection.DownloadLocations,
	}

	stepHash := models.StepHash{}

	stepsSpecDir := GetCollectionBaseDirPath(route)
	log.Debugln("  stepsSpecDir: ", stepsSpecDir)
	err := filepath.Walk(stepsSpecDir, func(path string, f os.FileInfo, err error) error {
		truncatedPath := strings.Replace(path, stepsSpecDir+"/", "", -1)
		match, matchErr := regexp.MatchString("([a-z]+).yml", truncatedPath)
		if matchErr != nil {
			return matchErr
		}

		if match {
			components := strings.Split(truncatedPath, "/")
			if len(components) == 4 {
				id := components[1]
				version := components[2]

				log.Debugf("Start parsing (StepId:%s) (Version:%s)", id, version)
				step, parseErr := parseStepYml(route.SteplibURI, path, id, version)
				if parseErr != nil {
					log.Debugf("  Failed to parse StepId: %v Version: %v", id, version)
					return parseErr
				}
				stepGroup, found := stepHash[id]
				if !found {
					stepGroup = models.StepGroupModel{
						Versions: map[string]models.StepModel{},
					}
				}
				stepGroup, err = addStepVersionToStepGroup(step, version, stepGroup)
				if err != nil {
					log.Debugf("  Failed to add step to step-group. (StepId:%v) (Version: %v) | Error: %v", id, version, err)
					return err
				}

				stepHash[id] = stepGroup
			} else {
				log.Debug("  * Path:", truncatedPath)
				log.Debug("  * Legth:", len(components))
			}
		}

		return err
	})
	if err != nil {
		log.Error("[STEPMAN] - Failed to walk through path:", err)
	}

	// log.Debugf("  collected steps: %#v\n", stepHash)
	collection.Steps = stepHash

	var bytes []byte
	// if DebugMode == true {
	bytes, err = json.MarshalIndent(collection, "", "\t")
	// } else {
	// 	bytes, err = json.Marshal(collection)
	// }
	if err != nil {
		log.Error("[STEPMAN] - Failed to parse json:", err)
		return []byte{}, err
	}

	return bytes, nil
}

// WriteStepSpecToFile ...
func WriteStepSpecToFile(templateCollection models.StepCollectionModel, route SteplibRoute) error {
	pth := GetStepSpecPath(route)

	if exist, err := pathutil.IsPathExists(pth); err != nil {
		log.Error("[STEPMAN] - Failed to check path:", err)
		return err
	} else if !exist {
		dir, _ := path.Split(pth)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	} else {
		err := os.Remove(pth)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(pth, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error("[STEPMAN] - Failed to close file:", err)
		}
	}()

	jsonContBytes, err := generateFormattedJSONForStepsSpec(route, templateCollection)
	if err != nil {
		return err
	}

	_, err = file.Write(jsonContBytes)
	if err != nil {
		return err
	}

	return nil
}

// ReadStepSpec ...
func ReadStepSpec(uri string) (models.StepCollectionModel, error) {
	log.Debugln("-> ReadStepSpec: ", uri)

	route, found := ReadRoute(uri)
	if !found {
		return models.StepCollectionModel{}, errors.New("No route found for lib: " + uri)
	}
	pth := GetStepSpecPath(route)
	file, err := os.Open(pth)
	if err != nil {
		return models.StepCollectionModel{}, err
	}

	var stepCollection models.StepCollectionModel
	parser := json.NewDecoder(file)
	if err = parser.Decode(&stepCollection); err != nil {
		return models.StepCollectionModel{}, err
	}
	return stepCollection, err
}

// DownloadAndUnZIP ...
func DownloadAndUnZIP(url, pth string) error {
	filePath := os.TempDir() + "step.zip"
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal("Failed to close file:", err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Fatal("Failed to remove file:", err)
		}
	}()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Fatal("Failed to close response body:", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		errorMsg := "Failed to download step.zip from: " + url
		return errors.New(errorMsg)
	}

	log.Info("Successfully downloaded step.zip")
	if _, err := io.Copy(file, response.Body); err != nil {
		return err
	}

	return unzip(filePath, pth)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Fatal(err)
				}
			}()

			if _, err = io.Copy(f, rc); err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		if err := extractAndWriteFile(f); err != nil {
			return err
		}
	}
	return nil
}
