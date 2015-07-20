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

	stepModel.ID = id
	stepModel.VersionTag = version
	stepModel.SteplibSource = collectionURI

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
func DownloadStep(step models.StepModel, collection models.StepCollectionModel) error {
	downloadLocations, err := collection.GetDownloadLocations(step)
	if err != nil {
		return err
	}

	stepPth := GetStepCacheDirPath(collection.SteplibSource, step)
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
			if err := DoGitClone(downloadLocation.Src, stepPth); err != nil {
				log.Errorf("[STEPMAN] - Failed to clone step (%s): %v", downloadLocation.Src, err)
			} else {
				success = true
				return nil
			}
		default:
			return fmt.Errorf("[STEPMAN] - Failed to download: Invalid download location (%#v) for step (%#v)", downloadLocation, step)
		}
	}

	if !success {
		return errors.New("Failed to download step")
	}
	return nil
}

// GetStepCacheDirPath ...
// Step's Cache dir path, where it's code lives.
func GetStepCacheDirPath(collectionURI string, step models.StepModel) string {
	return GetCacheBaseDir(collectionURI) + "/" + step.ID + "/" + step.VersionTag
}

// GetStepCollectionDirPath ...
// Step's Collection dir path, where it's spec (step.yml) lives.
func GetStepCollectionDirPath(collectionURI string, step models.StepModel) string {
	return GetCollectionBaseDirPath(collectionURI) + "/steps/" + step.ID + "/" + step.VersionTag
}

// semantic version (X.Y.Z)
// true if version 2 is greater then version 1
func isVersionGrater(version1, version2 string) bool {
	version1Slice := strings.Split(version1, ".")
	version2Slice := strings.Split(version2, ".")

	for i, num := range version1Slice {
		num1, err1 := strconv.ParseInt(num, 0, 64)
		if err1 != nil {
			log.Error("[STEPMAN] - Failed to parse int:", err1)
			return false
		}

		num2, err2 := strconv.ParseInt(version2Slice[i], 0, 64)
		if err2 != nil {
			log.Error("[STEPMAN] - Failed to parse int:", err2)
			return false
		}

		if num2 > num1 {
			return true
		}
	}
	return false
}

func addStepToStepGroup(step models.StepModel, stepGroup models.StepGroupModel) models.StepGroupModel {
	var newStepGroup models.StepGroupModel
	if len(stepGroup.Versions) > 0 {
		// Step Group already created -> new version of step
		newStepGroup = stepGroup

		if isVersionGrater(newStepGroup.Latest.VersionTag, step.VersionTag) {
			newStepGroup.Latest = step
		}
	} else {
		// Create Step Group
		newStepGroup = models.StepGroupModel{}
		newStepGroup.Latest = step
	}

	versions := make([]models.StepModel, len(newStepGroup.Versions))
	for idx, step := range newStepGroup.Versions {
		versions[idx] = step
	}
	versions = append(versions, step)

	newStepGroup.Versions = versions
	newStepGroup.ID = step.ID
	return newStepGroup
}

func generateFormattedJSONForStepsSpec(collectionURI string, templateCollection models.StepCollectionModel) ([]byte, error) {
	log.Debugln("-> generateFormattedJSONForStepsSpec")
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
			log.Debugf("  components: %#v\n", components)
			if len(components) == 4 {
				name := components[1]
				version := components[2]

				step, parseErr := parseStepYml(collectionURI, path, name, version)
				if parseErr != nil {
					return parseErr
				}
				stepGroup := addStepToStepGroup(step, stepHash[name])

				stepHash[name] = stepGroup
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

	log.Debugf("  collected steps: %#v\n", stepHash)
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
