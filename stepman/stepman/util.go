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
	"strconv"
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

	if err := stepModel.Validate(); err != nil {
		return models.StepModel{}, err
	}

	if err := stepModel.FillMissingDeafults(); err != nil {
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
func DownloadStep(collection models.StepCollectionModel, id, version string) error {
	log.Debugf("Download Step: %#v (%#v)\n", id, version)
	downloadLocations, err := collection.GetDownloadLocations(id, version)
	if err != nil {
		return err
	}

	stepPth := GetStepCacheDirPath(collection.SteplibSource, id, version)
	if exist, err := pathutil.IsPathExists(stepPth); err != nil {
		return err
	} else if exist {
		log.Info("[STEPMAN] - Step already downloaded")
		return nil
	}

	success := false
	for _, downloadLocation := range downloadLocations {
		switch downloadLocation.Type {
		case "zip":
			log.Info("[STEPMAN] - Downloading step from:", downloadLocation.Src)
			if err := DownloadAndUnZIP(downloadLocation.Src, stepPth); err != nil {
				log.Error("[STEPMAN] - Failed to download step.zip:", err)
			} else {
				success = true
				return nil
			}
		case "git":
			log.Info("[STEPMAN] - Git clone step from:", downloadLocation.Src)
			if err := DoGitCloneWithVersion(downloadLocation.Src, stepPth, version); err != nil {
				log.Errorf("[STEPMAN] - Failed to clone step (%s): %v", downloadLocation.Src, err)
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

// GetStepCacheDirPath ...
// Step's Cache dir path, where it's code lives.
func GetStepCacheDirPath(collectionURI string, id, version string) string {
	return GetCacheBaseDir(collectionURI) + "/" + id + "/" + version
}

// GetStepCollectionDirPath ...
// Step's Collection dir path, where it's spec (step.yml) lives.
func GetStepCollectionDirPath(collectionURI string, id, version string) string {
	return GetCollectionBaseDirPath(collectionURI) + "/steps/" + id + "/" + version
}

// semantic version (X.Y.Z)
// 1 if version 2 is greater then version 1, -1 if not
func compareVersions(version1, version2 string) int {
	version1Slice := strings.Split(version1, ".")
	version2Slice := strings.Split(version2, ".")

	for i, num := range version1Slice {
		num1, err1 := strconv.ParseInt(num, 0, 64)
		if err1 != nil {
			log.Error("[STEPMAN] - Failed to parse int:", err1)
			return 0
		}

		num2, err2 := strconv.ParseInt(version2Slice[i], 0, 64)
		if err2 != nil {
			log.Error("[STEPMAN] - Failed to parse int:", err2)
			return 0
		}

		if num2 > num1 {
			return 1
		}
	}
	return -1
}

func addStepVersionToStepGroup(step models.StepModel, version string, stepGroup models.StepGroupModel) models.StepGroupModel {
	if stepGroup.LatestVersionNumber != "" {
		if compareVersions(stepGroup.LatestVersionNumber, version) > 0 {
			stepGroup.LatestVersionNumber = version
		}
	} else {
		stepGroup.LatestVersionNumber = version
	}
	log.Debugf("SetGroup: %#v, versionParam: %#v, stepParam: %#v", stepGroup, version, step)
	stepGroup.Versions[version] = step
	return stepGroup
}

func generateFormattedJSONForStepsSpec(collectionURI string, templateCollection models.StepCollectionModel) ([]byte, error) {
	collection := models.StepCollectionModel{
		FormatVersion:        templateCollection.FormatVersion,
		GeneratedAtTimeStamp: time.Now().Unix(),
		SteplibSource:        collectionURI,
		DownloadLocations:    templateCollection.DownloadLocations,
	}

	stepHash := models.StepHash{}

	stepsSpecDir := GetCollectionBaseDirPath(collectionURI)
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
				step, parseErr := parseStepYml(collectionURI, path, id, version)
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
				stepGroup = addStepVersionToStepGroup(step, version, stepGroup)

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
func WriteStepSpecToFile(collectionURI string, templateCollection models.StepCollectionModel) error {
	pth := GetStepSpecPath(collectionURI)

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

	jsonContBytes, err := generateFormattedJSONForStepsSpec(collectionURI, templateCollection)
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
func ReadStepSpec(collectionURI string) (models.StepCollectionModel, error) {
	log.Debugln("-> ReadStepSpec: ", collectionURI)

	pth := GetStepSpecPath(collectionURI)
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
