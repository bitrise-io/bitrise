package stepman

import (
	"archive/zip"
	"encoding/json"
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
	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

const (
	// FormatVersion ...
	FormatVersion string = "0.9.0"
)

// DebugMode ...
var DebugMode bool

func parseStepYml(pth, id, version string) (models.StepModel, error) {
	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.StepModel{}, err
	}

	var stepJSON models.StepModel
	if err := yaml.Unmarshal(bytes, &stepJSON); err != nil {
		return models.StepModel{}, err
	}

	stepJSON.ID = id
	stepJSON.VersionTag = version
	stepJSON.StepLibSource = CollectionURI

	return stepJSON, nil
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

func generateFormattedJSONForStepsSpec() ([]byte, error) {
	collection := models.StepCollectionModel{
		FormatVersion:        FormatVersion,
		GeneratedAtTimeStamp: time.Now().Unix(),
		SteplibSource:        CollectionURI,
	}

	stepHash := models.StepHash{}

	stepsSpecDir := GetCurrentStepCollectionPath()
	err := filepath.Walk(stepsSpecDir, func(path string, f os.FileInfo, err error) error {
		truncatedPath := strings.Replace(path, stepsSpecDir, "", -1)
		match, matchErr := regexp.MatchString("([a-z]+).yml", truncatedPath)
		if matchErr != nil {
			return matchErr
		}

		if match {
			components := strings.Split(truncatedPath, "/")
			if len(components) == 4 {
				name := components[1]
				version := components[2]

				step, parseErr := parseStepYml(path, name, version)
				if parseErr != nil {
					return parseErr
				}
				stepGroup := addStepToStepGroup(step, stepHash[name])

				stepHash[name] = stepGroup
			} else {
				log.Debug("[STEPMAN] - Path:", truncatedPath)
				log.Debug("[STEPMAN] - Legth:", len(components))
			}
		}

		return err
	})
	if err != nil {
		log.Error("[STEPMAN] - Failed to walk through path:", err)
	}

	collection.Steps = stepHash

	var bytes []byte
	if DebugMode == true {
		bytes, err = json.MarshalIndent(collection, "", "\t")
	} else {
		bytes, err = json.Marshal(collection)
	}
	if err != nil {
		log.Error("[STEPMAN] - Failed to parse json:", err)
		return []byte{}, err
	}

	return bytes, nil
}

// WriteStepSpecToFile ...
func WriteStepSpecToFile() error {
	pth := GetCurrentStepSpecPath()

	if exist, err := pathutil.IsPathExists(pth); err != nil {
		log.Error("[STEPMAN] - Failed to check path:", err)
		return err
	} else if exist == false {
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

	jsonContBytes, err := generateFormattedJSONForStepsSpec()
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
func ReadStepSpec() (models.StepCollectionModel, error) {
	pth := GetCurrentStepSpecPath()
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

// DownloadStep ...
func DownloadStep(step models.StepModel) error {
	primarySource := step.PrimaryDownloadLocation()
	log.Info("Primary source:", primarySource)

	destPath := GetStepPath(step)
	filePath := os.TempDir() + "step.zip"
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() error {
		if err := file.Close(); err != nil {
			return err
		}
		return os.Remove(filePath)
	}()

	response, err := http.Get(primarySource)
	if err != nil {
		return err
	}
	defer func() error {
		return response.Body.Close()
	}()

	if response.StatusCode == http.StatusOK {
		log.Info("Successfully downloaded step.zip")
		if _, err := io.Copy(file, response.Body); err != nil {
			return err
		}

		return unzip(filePath, destPath)
	}
	log.Info("Failed to download step.zip:", response.StatusCode)

	gitSource := step.Source["git"]
	pth := GetStepPath(step)
	return DoGitUpdate(gitSource, pth)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetStepPath ...
func GetStepPath(step models.StepModel) string {
	return GetCurrentStepCahceDir() + step.ID + "/" + step.VersionTag + "/"
}
