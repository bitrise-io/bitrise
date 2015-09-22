package stepman

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/urlutil"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

// DebugMode ...
var DebugMode bool

// ParseStepYml ...
func ParseStepYml(pth string, validate bool) (models.StepModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
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

	if validate {
		if err := stepModel.Validate(true); err != nil {
			return models.StepModel{}, err
		}
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return models.StepModel{}, err
	}

	return stepModel, nil
}

// ParseStepCollection ...
func ParseStepCollection(pth string) (models.StepCollectionModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
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
			log.Debug("[STEPMAN] - Downloading step from: ", downloadLocation.Src)
			if err := cmdex.DownloadAndUnZIP(downloadLocation.Src, stepPth); err != nil {
				log.Warn("[STEPMAN] - Failed to download step.zip: ", err)
			} else {
				success = true
				return nil
			}
		case "git":
			log.Debug("[STEPMAN] - Git clone step from: ", downloadLocation.Src)
			if err := cmdex.GitCloneTagOrBranchAndValidateCommitHash(downloadLocation.Src, stepPth, version, commithash); err != nil {
				log.Warnf("[STEPMAN] - Failed to clone step (%s): %v", downloadLocation.Src, err)
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
		r, err := versions.CompareVersions(stepGroup.LatestVersionNumber, version)
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

func generateStepLib(route SteplibRoute, templateCollection models.StepCollectionModel) (models.StepCollectionModel, error) {
	collection := models.StepCollectionModel{
		FormatVersion:         templateCollection.FormatVersion,
		GeneratedAtTimeStamp:  time.Now().Unix(),
		SteplibSource:         templateCollection.SteplibSource,
		DownloadLocations:     templateCollection.DownloadLocations,
		AssetsDownloadBaseURI: templateCollection.AssetsDownloadBaseURI,
	}

	stepHash := models.StepHash{}

	stepsSpecDir := GetCollectionBaseDirPath(route)
	log.Debugln("  stepsSpecDir: ", stepsSpecDir)
	err := filepath.Walk(stepsSpecDir, func(pth string, f os.FileInfo, err error) error {
		truncatedPath := strings.Replace(pth, stepsSpecDir+"/", "", -1)
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
				step, parseErr := ParseStepYml(pth, true)
				if parseErr != nil {
					log.Debugf("  Failed to parse StepId: %v Version: %v", id, version)
					return parseErr
				}

				// Check for assets
				if collection.AssetsDownloadBaseURI != "" {
					assetsFolderPth := path.Join(stepsSpecDir, components[0], components[1], "assets")
					exist, err := pathutil.IsPathExists(assetsFolderPth)
					if err != nil {
						return err
					}
					if exist {
						assetsMap := map[string]string{}
						err := filepath.Walk(assetsFolderPth, func(pth string, f os.FileInfo, err error) error {
							_, file := filepath.Split(pth)
							if pth != assetsFolderPth && file != "" {
								assetURI, err := urlutil.Join(collection.AssetsDownloadBaseURI, id, "assets", file)
								if err != nil {
									return err
								}
								assetsMap[file] = assetURI
							}
							return nil
						})

						if err != nil {
							log.Debugf("  Failed to add assets, at (%s) | Error: %v", assetsFolderPth, err)
							return err
						}

						step.AssetURLs = assetsMap
					}
				}

				// Add to stepgroup
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
		return models.StepCollectionModel{}, err
	}
	collection.Steps = stepHash
	return collection, nil
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

	collection, err := generateStepLib(route, templateCollection)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(collection, "", "\t")
	if err != nil {
		return err
	}
	return fileutil.WriteBytesToFile(pth, bytes)
}

// ReadStepSpec ...
func ReadStepSpec(uri string) (models.StepCollectionModel, error) {
	log.Debugln("-> ReadStepSpec: ", uri)

	route, found := ReadRoute(uri)
	if !found {
		return models.StepCollectionModel{}, errors.New("No route found for lib: " + uri)
	}
	pth := GetStepSpecPath(route)
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return models.StepCollectionModel{}, err
	}
	var stepLib models.StepCollectionModel
	if err := json.Unmarshal(bytes, &stepLib); err != nil {
		return models.StepCollectionModel{}, err
	}
	return stepLib, nil
}

// ReGenerateStepSpec ...
func ReGenerateStepSpec(route SteplibRoute) error {
	pth := GetCollectionBaseDirPath(route)
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if !exists {
		return errors.New("[STEPMAN] - Not initialized")
	}

	specPth := pth + "/steplib.yml"
	collection, err := ParseStepCollection(specPth)
	if err != nil {
		return err
	}

	if err := WriteStepSpecToFile(collection, route); err != nil {
		return err
	}
	return nil
}
