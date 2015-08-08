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
	// StepManDirPath ...
	StepManDirPath string
	// CollectionsDirPath ...
	CollectionsDirPath string

	routingFilePath string
)

// SteplibRoute ...
type SteplibRoute struct {
	SteplibURI  string
	FolderAlias string
}

// SteplibRoutes ...
type SteplibRoutes []SteplibRoute

// GetRoute ...
func (routes SteplibRoutes) GetRoute(URI string) (route SteplibRoute, found bool) {
	for _, route := range routes {
		if route.SteplibURI == URI {
			return route, true
		}
	}
	return SteplibRoute{}, false
}

// ReadRoute ...
func ReadRoute(uri string) (route SteplibRoute, found bool) {
	routes, err := readRouteMap()
	if err != nil {
		return SteplibRoute{}, false
	}

	return routes.GetRoute(uri)
}

func (routes SteplibRoutes) writeToFile() error {
	if exist, err := pathutil.IsPathExists(StepManDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(StepManDirPath, 0777); err != nil {
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

	routeMap := map[string]string{}
	for _, route := range routes {
		routeMap[route.SteplibURI] = route.FolderAlias
	}

	bytes, err := json.MarshalIndent(routeMap, "", "\t")
	if err != nil {
		log.Error("[STEPMAN] - Failed to parse json:", err)
		return err
	}

	if _, err := file.Write(bytes); err != nil {
		return err
	}
	return nil
}

// RemoveDir ...
func RemoveDir(dirPth string) error {
	if exist, err := pathutil.IsPathExists(dirPth); err != nil {
		return err
	} else if exist {
		if err := os.RemoveAll(dirPth); err != nil {
			return err
		}
	}
	return nil
}

// CleanupRoute ...
func CleanupRoute(route SteplibRoute) error {
	pth := CollectionsDirPath + "/" + route.FolderAlias
	if err := RemoveDir(pth); err != nil {
		return err
	}
	if err := RemoveRoute(route); err != nil {
		return err
	}
	return nil
}

// RootExistForCollection ...
func RootExistForCollection(collectionURI string) (bool, error) {
	routes, err := readRouteMap()
	if err != nil {
		return false, err
	}

	_, found := routes.GetRoute(collectionURI)
	return found, nil
}

func getAlias(uri string) (string, error) {
	routes, err := readRouteMap()
	if err != nil {
		return "", err
	}

	route, found := routes.GetRoute(uri)
	if found == false {
		return "", errors.New("No routes exist for uri:" + uri)
	}
	return route.FolderAlias, nil
}

// RemoveRoute ...
func RemoveRoute(route SteplibRoute) error {
	routes, err := readRouteMap()
	if err != nil {
		return err
	}

	newRoutes := SteplibRoutes{}
	for _, aRoute := range routes {
		if aRoute.SteplibURI != route.SteplibURI {
			newRoutes = append(newRoutes, aRoute)
		}
	}
	if err := newRoutes.writeToFile(); err != nil {
		return err
	}
	return nil
}

// AddRoute ...
func AddRoute(route SteplibRoute) error {
	routes, err := readRouteMap()
	if err != nil {
		return err
	}

	routes = append(routes, route)
	if err := routes.writeToFile(); err != nil {
		return err
	}

	return nil
}

// GenerateFolderAlias ...
func GenerateFolderAlias() string {
	return fmt.Sprintf("%v", time.Now().Unix())
}

func readRouteMap() (SteplibRoutes, error) {
	if exist, err := pathutil.IsPathExists(routingFilePath); err != nil {
		return SteplibRoutes{}, err
	} else if !exist {
		return SteplibRoutes{}, nil
	}

	file, e := os.Open(routingFilePath)
	if e != nil {
		return SteplibRoutes{}, e
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("[STEPMAN] - Failed to close file:", err)
		}
	}()

	var routeMap map[string]string
	parser := json.NewDecoder(file)
	if err := parser.Decode(&routeMap); err != nil {
		return SteplibRoutes{}, err
	}

	routes := []SteplibRoute{}
	for key, value := range routeMap {
		routes = append(routes, SteplibRoute{
			SteplibURI:  key,
			FolderAlias: value,
		})
	}

	return routes, nil
}

// CreateStepManDirIfNeeded ...
func CreateStepManDirIfNeeded() error {
	if exist, err := pathutil.IsPathExists(StepManDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(StepManDirPath, 0777); err != nil {
			return err
		}
	}
	return nil
}

// GetStepSpecPath ...
func GetStepSpecPath(route SteplibRoute) string {
	return CollectionsDirPath + "/" + route.FolderAlias + "/spec/spec.json"
}

// GetCacheBaseDir ...
func GetCacheBaseDir(route SteplibRoute) string {
	return CollectionsDirPath + "/" + route.FolderAlias + "/cache"
}

// GetCollectionBaseDirPath ...
func GetCollectionBaseDirPath(route SteplibRoute) string {
	return CollectionsDirPath + "/" + route.FolderAlias + "/collection"
}

// GetAllStepCollectionPath ...
func GetAllStepCollectionPath() []string {
	routes, err := readRouteMap()
	if err != nil {
		log.Error("[STEPMAN] - Failed to read step specs path:", err)
		return []string{}
	}

	sources := []string{}
	for _, route := range routes {
		sources = append(sources, route.SteplibURI)
	}

	return sources
}

// GetStepCacheDirPath ...
// Step's Cache dir path, where it's code lives.
func GetStepCacheDirPath(route SteplibRoute, id, version string) string {
	return GetCacheBaseDir(route) + "/" + id + "/" + version
}

// GetStepCollectionDirPath ...
// Step's Collection dir path, where it's spec (step.yml) lives.
func GetStepCollectionDirPath(route SteplibRoute, id, version string) string {
	return GetCollectionBaseDirPath(route) + "/steps/" + id + "/" + version
}

// Life cycle
func init() {
	StepManDirPath = pathutil.UserHomeDir() + "/" + StepmanDirname
	routingFilePath = StepManDirPath + "/" + RoutingFilename
	CollectionsDirPath = StepManDirPath + "/" + CollectionsDirname
}
