package stepman

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
)

const (
	// StepmanDirname ...
	StepmanDirname string = ".stepman"
	// RoutingFilename ...
	RoutingFilename string = "routing.json"
	// CollectionsDirname ...
	CollectionsDirname string = "step_collections"
)

var (
	stepManDirPath  string
	routingFilePath string

	// CollectionURI ...
	CollectionURI string

	// CollectionsDirPath ...
	CollectionsDirPath string
)

// RouteMap ...
type RouteMap map[string]string

// RootExistForCurrentCollection ...
func RootExistForCurrentCollection() (bool, error) {
	RouteMap, err := readRouteMap()
	if err != nil {
		return false, err
	}

	if RouteMap[CollectionURI] != "" {
		return true, nil
	}
	return false, nil
}

func getAlias(source string) (string, error) {
	routeMap, err := readRouteMap()
	if err != nil {
		return "", err
	}

	if routeMap[source] == "" {
		return "", errors.New("No route found for source")
	}

	return routeMap[source], nil
}

func addRoute(source, alias string) error {
	RouteMap, err := readRouteMap()
	if err != nil {
		return err
	}

	if RouteMap[source] != "" {
		return errors.New("Route already exist for source")
	}

	RouteMap[source] = alias

	if err := writeRouteMapToFile(RouteMap); err != nil {
		return err
	}

	return nil
}

func generateFolderAlias(source string) string {
	return fmt.Sprintf("%v", time.Now().Unix())
}

func writeRouteMapToFile(RouteMap RouteMap) error {
	if exist, err := pathutil.IsPathExists(stepManDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(stepManDirPath, 0777); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(routingFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("[STEPMAN] - Failed to close file:", err)
		}
	}()

	bytes, err := json.MarshalIndent(RouteMap, "", "\t")
	if err != nil {
		log.Error("[STEPMAN] - Failed to parse json:", err)
		return err
	}

	if _, err := file.Write(bytes); err != nil {
		return err
	}
	return nil
}

func readRouteMap() (RouteMap, error) {
	if exist, err := pathutil.IsPathExists(routingFilePath); err != nil {
		return RouteMap{}, err
	} else if !exist {
		return RouteMap{}, nil
	}

	file, e := os.Open(routingFilePath)
	if e != nil {
		return RouteMap{}, e
	}

	var routeMap RouteMap
	parser := json.NewDecoder(file)
	if err := parser.Decode(&routeMap); err != nil {
		return RouteMap{}, err
	}
	return routeMap, nil
}

// CreateStepManDirIfNeeded ...
func CreateStepManDirIfNeeded() error {
	if exist, err := pathutil.IsPathExists(stepManDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(stepManDirPath, 0777); err != nil {
			return err
		}
	}
	return nil
}

// SetupCurrentRouting ...
func SetupCurrentRouting() error {
	if CollectionURI == "" {
		return errors.New("No collection path defined")
	}

	alias := generateFolderAlias(CollectionURI)
	return addRoute(CollectionURI, alias)
}

// GetCurrentStepSpecPath ...
func GetCurrentStepSpecPath() string {
	alias, err := getAlias(CollectionURI)
	if err != nil {
		log.Error("[STEPMAN] - Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDirPath + alias + "/spec/spec.json"
}

// GetCurrentStepCacheDir ...
func GetCurrentStepCacheDir() string {
	alias, err := getAlias(CollectionURI)
	if err != nil {
		log.Error("[STEPMAN] - Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDirPath + alias + "/cache/"
}

// GetCurrentStepCollectionPath ...
func GetCurrentStepCollectionPath() string {
	alias, err := getAlias(CollectionURI)
	if err != nil {
		log.Error("[STEPMAN] - Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDirPath + alias + "/collection/"
}

// Life cycle
func init() {
	stepManDirPath = pathutil.UserHomeDir() + "/" + StepmanDirname + "/"
	routingFilePath = stepManDirPath + RoutingFilename
	CollectionsDirPath = stepManDirPath + CollectionsDirname + "/"
}
