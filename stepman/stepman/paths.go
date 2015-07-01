package stepman

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bitrise-io/go-pathutil"
)

const (
	STEPLIB_SOURCE      string = "https://github.com/steplib/steplib"
	STEP_COLLECTION_GIT string = "https://github.com/steplib/steplib.git"

	STEPMAN_DIR            string = "/.stepman/"
	ROUTING_PTH_SUFFIX     string = "routing.json"
	COLLECTIONS_DIR_SUFFIX string = "step_collections/"
)

var (
	stepManDir    string
	routeFilePath string

	CollectionPath string

	CollectionsDir string
)

type RouteMap map[string]string

func (route RouteMap) getFirstKey() string {
	for key, _ := range route {
		return key
	}
	return ""
}

func (route RouteMap) getFirstValue() string {
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
	routeMap, err := readRouteMap()
	if err != nil {
		return err
	}

	if routeMap[route.getFirstKey()] != "" {
		return errors.New("Route already exist for source")
	}

	routeMap[route.getFirstKey()] = route[route.getFirstKey()]

	err = writeRouteMapToFile(routeMap)
	if err != nil {
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

func writeRouteMapToFile(routeMap RouteMap) error {
	exist, err := pathutil.IsPathExists(stepManDir)
	if err != nil {
		return err
	}
	if exist == false {
		err := os.MkdirAll(stepManDir, 0777)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(routeFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println("Failed to close file: %s", err)
		}
	}()

	bytes, err := json.MarshalIndent(routeMap, "", "\t")
	//bytes, err := json.Marshal(routeMap)
	if err != nil {
		fmt.Println("error:", err)
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func readRouteMap() (RouteMap, error) {
	exist, err := pathutil.IsPathExists(routeFilePath)
	if err != nil {
		return RouteMap{}, err
	}
	if exist == false {
		return RouteMap{}, nil
	}

	file, e := os.Open(routeFilePath)
	if e != nil {
		return RouteMap{}, e
	}

	var routeMap RouteMap
	parser := json.NewDecoder(file)
	if err = parser.Decode(&routeMap); err != nil {
		return RouteMap{}, err
	}
	return routeMap, nil
}

// Interface
func CreateStepManDirIfNeeded() error {
	exist, err := pathutil.IsPathExists(stepManDir)
	if err != nil {
		return err
	}
	if exist == false {
		err := os.MkdirAll(stepManDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetupCurrentRouting() error {
	if CollectionPath == "" {
		return errors.New("No collection path defined")
	}

	route := generateRoute(CollectionPath)
	return addRoute(route)
}

func GetCurrentStepSpecPath() string {
	route, err := getRoute(CollectionPath)
	if err != nil {
		fmt.Println("Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDir + route.getFirstValue() + "/spec/spec.json"
}

func GetCurrentStepCahceDir() string {
	route, err := getRoute(CollectionPath)
	if err != nil {
		fmt.Println("Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDir + route.getFirstValue() + "/cache/"
}

func GetCurrentStepCollectionPath() string {
	route, err := getRoute(CollectionPath)
	if err != nil {
		fmt.Println("Failed to generate current step spec path:", err)
		return ""
	}
	return CollectionsDir + route.getFirstValue() + "/collection/"
}

// Life cycle
func init() {
	stepManDir = pathutil.UserHomeDir() + STEPMAN_DIR
	routeFilePath = stepManDir + ROUTING_PTH_SUFFIX
	CollectionsDir = stepManDir + COLLECTIONS_DIR_SUFFIX
}
