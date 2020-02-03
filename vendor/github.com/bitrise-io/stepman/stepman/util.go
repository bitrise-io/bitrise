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

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/urlutil"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

// ParseStepGroupInfoModel ...
func ParseStepGroupInfoModel(pth string) (models.StepGroupInfoModel, bool, error) {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return models.StepGroupInfoModel{}, false, err
	} else if !exist {
		return models.StepGroupInfoModel{}, false, nil
	}

	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return models.StepGroupInfoModel{}, true, err
	}

	var globalStepInfo models.StepGroupInfoModel
	if err := yaml.Unmarshal(bytes, &globalStepInfo); err != nil {
		return models.StepGroupInfoModel{}, true, err
	}

	return globalStepInfo, true, nil
}

// ParseStepDefinition ...
func ParseStepDefinition(pth string, validate bool) (models.StepModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return models.StepModel{}, err
	}

	return parseStepModel(bytes, validate)
}

func parseStepModel(bytes []byte, validate bool) (models.StepModel, error) {
	var stepModel models.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		return models.StepModel{}, err
	}

	if err := stepModel.Normalize(); err != nil {
		return models.StepModel{}, err
	}

	if validate {
		if err := stepModel.Audit(); err != nil {
			return models.StepModel{}, err
		}
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return models.StepModel{}, err
	}

	return stepModel, nil
}

// ParseStepGroupInfo ...
func ParseStepGroupInfo(pth string) (models.StepGroupInfoModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return models.StepGroupInfoModel{}, err
	}

	var stepGroupInfo models.StepGroupInfoModel
	if err := yaml.Unmarshal(bytes, &stepGroupInfo); err != nil {
		return models.StepGroupInfoModel{}, err
	}

	return stepGroupInfo, nil
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
func DownloadStep(collectionURI string, collection models.StepCollectionModel, id, version, commithash string) error {
	downloadLocations, err := collection.GetDownloadLocations(id, version)
	if err != nil {
		return err
	}

	route, found := ReadRoute(collectionURI)
	if !found {
		return fmt.Errorf("No routing found for lib: %s", collectionURI)
	}

	stepPth := GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepPth); err != nil {
		return err
	} else if exist {
		return nil
	}

	success := false
	for _, downloadLocation := range downloadLocations {
		switch downloadLocation.Type {
		case "zip":
			err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
				return command.DownloadAndUnZIP(downloadLocation.Src, stepPth)
			})

			if err != nil {
				log.Warnf("Failed to download step.zip: %s", err)
			} else {
				success = true
				return nil
			}
		case "git":
			err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
				repo, err := git.New(stepPth)
				if err != nil {
					return err
				}

				if err := repo.CloneTagOrBranch(downloadLocation.Src, version).Run(); err != nil {
					return err
				}

				hash, err := repo.RevParse("HEAD").RunAndReturnTrimmedCombinedOutput()
				if err != nil {
					return err
				}

				if hash != commithash {
					return fmt.Errorf("commit hash (%s) doesn't match the one specified (%s) for the version tag (%s)", hash, commithash, version)
				}
				return nil
			})

			if err != nil {
				log.Warnf("Failed to clone step (%s): %v", downloadLocation.Src, err)
			} else {
				success = true
				return nil
			}
		default:
			return fmt.Errorf("Failed to download: Invalid download location (%#v) for step %#v (%#v)", downloadLocation, id, version)
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

	stepsSpecDirPth := GetLibraryBaseDirPath(route)
	if err := filepath.Walk(stepsSpecDirPth, func(pth string, f os.FileInfo, err error) error {
		truncatedPath := strings.Replace(pth, stepsSpecDirPth+"/", "", -1)
		match, matchErr := regexp.MatchString("([a-z]+).yml", truncatedPath)
		if matchErr != nil {
			return matchErr
		}

		if match {
			components := strings.Split(truncatedPath, "/")
			if len(components) == 4 {
				stepsDirName := components[0]
				stepID := components[1]
				stepVersion := components[2]

				step, parseErr := ParseStepDefinition(pth, true)
				if parseErr != nil {
					return parseErr
				}

				stepGroupInfo := models.StepGroupInfoModel{}

				// Check for step-info.yml - STEP_SPEC_DIR/steps/step-id/step-info.yml
				stepGroupInfoPth := filepath.Join(stepsSpecDirPth, stepsDirName, stepID, "step-info.yml")
				if exist, err := pathutil.IsPathExists(stepGroupInfoPth); err != nil {
					return err
				} else if exist {
					deprecationInfo, err := ParseStepGroupInfo(stepGroupInfoPth)
					if err != nil {
						return err
					}

					stepGroupInfo.RemovalDate = deprecationInfo.RemovalDate
					stepGroupInfo.DeprecateNotes = deprecationInfo.DeprecateNotes
					stepGroupInfo.Maintainer = deprecationInfo.Maintainer
				}

				// Check for assets - STEP_SPEC_DIR/steps/step-id/assets
				if collection.AssetsDownloadBaseURI != "" {
					assetsFolderPth := path.Join(stepsSpecDirPth, stepsDirName, stepID, "assets")
					exist, err := pathutil.IsPathExists(assetsFolderPth)
					if err != nil {
						return err
					}
					if exist {
						assetsMap := map[string]string{}
						err := filepath.Walk(assetsFolderPth, func(pth string, f os.FileInfo, err error) error {
							_, file := filepath.Split(pth)
							if pth != assetsFolderPth && file != "" {
								assetURI, err := urlutil.Join(collection.AssetsDownloadBaseURI, stepID, "assets", file)
								if err != nil {
									return err
								}
								assetsMap[file] = assetURI
							}
							return nil
						})

						if err != nil {
							return err
						}

						step.AssetURLs = assetsMap
						stepGroupInfo.AssetURLs = assetsMap
					}
				}

				// Add to stepgroup
				stepGroup, found := stepHash[stepID]
				if !found {
					stepGroup = models.StepGroupModel{
						Versions: map[string]models.StepModel{},
					}
				}
				stepGroup, err = addStepVersionToStepGroup(step, stepVersion, stepGroup)
				if err != nil {
					return err
				}

				stepGroup.Info = stepGroupInfo

				stepHash[stepID] = stepGroup
			} else {
			}
		}

		return err
	}); err != nil {
		return models.StepCollectionModel{}, fmt.Errorf("Failed to walk through path, error: %s", err)
	}

	collection.Steps = stepHash

	return collection, nil
}

func generateSlimStepLib(collection models.StepCollectionModel) models.StepCollectionModel {

	slimCollection := models.StepCollectionModel{
		FormatVersion:         collection.FormatVersion,
		GeneratedAtTimeStamp:  collection.GeneratedAtTimeStamp,
		SteplibSource:         collection.SteplibSource,
		DownloadLocations:     collection.DownloadLocations,
		AssetsDownloadBaseURI: collection.AssetsDownloadBaseURI,
	}
	steps := models.StepHash{}

	for stepID, stepGroupModel := range collection.Steps {
		steps[stepID] = models.StepGroupModel{
			Info:     stepGroupModel.Info,
			Versions: map[string]models.StepModel{stepGroupModel.LatestVersionNumber: stepGroupModel.Versions[stepGroupModel.LatestVersionNumber]},
		}
	}

	slimCollection.Steps = steps

	return slimCollection
}

// WriteStepSpecToFile ...
func WriteStepSpecToFile(templateCollection models.StepCollectionModel, route SteplibRoute) error {
	pth := GetStepSpecPath(route)

	if exist, err := pathutil.IsPathExists(pth); err != nil {
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

	if err := fileutil.WriteBytesToFile(pth, bytes); err != nil {
		return err
	}

	pth = GetSlimStepSpecPath(route)
	slimCollection := generateSlimStepLib(collection)
	if err != nil {
		return err
	}

	bytes, err = json.MarshalIndent(slimCollection, "", "\t")
	if err != nil {
		return err
	}

	return fileutil.WriteBytesToFile(pth, bytes)
}

// ReadStepSpec ...
func ReadStepSpec(uri string) (models.StepCollectionModel, error) {
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

// ReadStepVersionInfo ...
func ReadStepVersionInfo(collectionURI, stepID, stepVersionID string) (models.StepVersionModel, error) {
	// Input validation
	if stepID == "" {
		return models.StepVersionModel{}, errors.New("Missing required input: step id")
	}

	// Check if step exist in collection
	collection, err := ReadStepSpec(collectionURI)
	if err != nil {
		return models.StepVersionModel{}, fmt.Errorf("Failed to read steps spec (spec.json), err: %s", err)
	}

	stepWithVersion, stepFound, versionFound := collection.GetStepVersion(stepID, stepVersionID)
	if !stepFound {
		return models.StepVersionModel{}, fmt.Errorf("Collection doesn't contain step with id: %s", stepID)
	} else if !versionFound {
		return models.StepVersionModel{}, fmt.Errorf("Collection doesn't contain step (%s) with version: %s", stepID, stepVersionID)
	}

	return stepWithVersion, nil
}

// ReGenerateLibrarySpec ...
func ReGenerateLibrarySpec(route SteplibRoute) error {
	pth := GetLibraryBaseDirPath(route)
	if exists, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if !exists {
		return errors.New("Not initialized")
	}

	specPth := pth + "/steplib.yml"
	collection, err := ParseStepCollection(specPth)
	if err != nil {
		return err
	}

	return WriteStepSpecToFile(collection, route)
}
