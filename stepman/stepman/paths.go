package stepman

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil"
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

func (route RouteMap) getSingleKey() string {
	for key := range route {
		return key
	}
	return ""
}

func (route RouteMap) getSingleValue() string {
	for _, value := range route {
		return value
	}
	return ""
}

func getRoute(source string) (RouteMap, error) {
	routeMap, err := readRouteMap()
	if err != nil {
		return RouteMap{}, err
	}

	if routeMap[source] == "" {
		return RouteMap{}, errors.New("No route found for source")
	}

	r := RouteMap{
		source: routeMap[source],
	}

	return r, nil
}

func addRoute(route RouteMap) error {
	RouteMap, err := readRouteMap()
	if err != nil {
		return err
	}

	if RouteMap[route.getSingleKey()] != "" {
		return errors.New("Route already exist for source")
	}

	RouteMap[route.getSingleKey()] = route[route.getSingleKey()]

	if err := writeRouteMapToFile(RouteMap); err != nil {
		return err
	}

	return nil
}

func generateRoute(source string) RouteMap {
	timeStamp := fmt.Sprintf("%v", time.Now().Unix())
	return RouteMap{
		source: timeStamp,
	}
}

func writeRouteMapToFile(RouteMap RouteMap) error {

	if exist, err := pathutil.IsPathExists(stepManDirPath); err != nil {
		return err
	} else if exist == false {
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
	} else if exist == false {
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
	} else if exist == false {
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

	route := generateRoute(CollectionURI)
	return addRoute(route)
}

// GetCurrentStepSpecPath ...
func GetCurrentStepSpecPath() string {
	route, err := getRoute(CollectionURI)
	if err != nil {
		log.Error("[STEPMAN] - Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDirPath + route.getSingleValue() + "/spec/spec.json"
}

// GetCurrentStepCahceDir ...
func GetCurrentStepCahceDir() string {
	route, err := getRoute(CollectionURI)
	if err != nil {
		log.Error("[STEPMAN] - Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDirPath + route.getSingleValue() + "/cache/"
}

// GetCurrentStepCollectionPath ...
func GetCurrentStepCollectionPath() string {
	route, err := getRoute(CollectionURI)
	if err != nil {
		log.Error("[STEPMAN] - Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDirPath + route.getSingleValue() + "/collection/"
}

// Life cycle
func init() {
	stepManDirPath = pathutil.UserHomeDir() + "/" + StepmanDirname + "/"
	routingFilePath = stepManDirPath + RoutingFilename
	CollectionsDirPath = stepManDirPath + CollectionsDirname + "/"
}
