package stepman

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	// StepmanDirname ...
	StepmanDirname = ".stepman"
	// RoutingFilename ...
	RoutingFilename = "routing.json"
	// CollectionsDirname ...
	CollectionsDirname = "step_collections"
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
			pth := filepath.Join(GetCollectionsDirPath(), route.FolderAlias)
			exist, err := pathutil.IsPathExists(pth)
			if err != nil {
				return SteplibRoute{}, false
			} else if !exist {
				return SteplibRoute{}, false
			}
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
	routeMap := map[string]string{}
	for _, route := range routes {
		routeMap[route.SteplibURI] = route.FolderAlias
	}
	bytes, err := json.MarshalIndent(routeMap, "", "\t")
	if err != nil {
		return err
	}
	return fileutil.WriteBytesToFile(getRoutingFilePath(), bytes)
}

// CleanupRoute ...
func CleanupRoute(route SteplibRoute) error {
	pth := filepath.Join(GetCollectionsDirPath(), route.FolderAlias)
	if err := command.RemoveDir(pth); err != nil {
		return err
	}
	return RemoveRoute(route)
}

// CleanupDanglingLibrary ...
func CleanupDanglingLibrary(URI string) error {
	route := SteplibRoute{
		SteplibURI:  URI,
		FolderAlias: "",
	}
	return RemoveRoute(route)
}

// RootExistForLibrary ...
func RootExistForLibrary(collectionURI string) (bool, error) {
	routes, err := readRouteMap()
	if err != nil {
		return false, err
	}

	_, found := routes.GetRoute(collectionURI)
	return found, nil
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
	return newRoutes.writeToFile()
}

// AddRoute ...
func AddRoute(route SteplibRoute) error {
	routes, err := readRouteMap()
	if err != nil {
		return err
	}

	routes = append(routes, route)
	return routes.writeToFile()
}

// GenerateFolderAlias ...
func GenerateFolderAlias() string {
	return fmt.Sprintf("%v", time.Now().Unix())
}

func readRouteMap() (SteplibRoutes, error) {
	exist, err := pathutil.IsPathExists(getRoutingFilePath())
	if err != nil {
		return SteplibRoutes{}, err
	} else if !exist {
		return SteplibRoutes{}, nil
	}

	bytes, err := fileutil.ReadBytesFromFile(getRoutingFilePath())
	if err != nil {
		return SteplibRoutes{}, err
	}
	var routeMap map[string]string
	if err := json.Unmarshal(bytes, &routeMap); err != nil {
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
	return os.MkdirAll(GetStepmanDirPath(), 0777)
}

// GetStepSpecPath ...
func GetStepSpecPath(route SteplibRoute) string {
	return filepath.Join(GetCollectionsDirPath(), route.FolderAlias, "spec", "spec.json")
}

// GetSlimStepSpecPath ...
func GetSlimStepSpecPath(route SteplibRoute) string {
	return filepath.Join(GetCollectionsDirPath(), route.FolderAlias, "spec", "slim-spec.json")
}

// GetCacheBaseDir ...
func GetCacheBaseDir(route SteplibRoute) string {
	return filepath.Join(GetCollectionsDirPath(), route.FolderAlias, "cache")
}

// GetLibraryBaseDirPath ...
func GetLibraryBaseDirPath(route SteplibRoute) string {
	return filepath.Join(GetCollectionsDirPath(), route.FolderAlias, "collection")
}

// GetStepCollectionSpecPath Location of steplib.yml of the collection marked by this route
func GetStepCollectionSpecPath(route SteplibRoute) string {
	return filepath.Join(GetLibraryBaseDirPath(route), "steplib.yml")
}

// GetAllStepCollectionPath ...
func GetAllStepCollectionPath() []string {
	routes, err := readRouteMap()
	if err != nil {
		log.Errorf("Failed to read step specs path, error: %s", err)
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
	return filepath.Join(GetCacheBaseDir(route), id, version)
}

// GetStepBinDirPath stores the precompiled binaries of the step (compressed and uncompressed)
func GetStepBinDirPath(route SteplibRoute, id string) string {
	return filepath.Join(GetCacheBaseDir(route), id, "binrepo")
}

func GetStepBinDirPathForVersion(route SteplibRoute, id, version string) string {
	return filepath.Join(GetStepBinDirPath(route, id), version)
}

func executableBinaryName() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

func executableChecksumName() string {
	return "sha256sum"
}

func binaryPatchName(patchFromVersion string) string {
	return fmt.Sprintf("%s-%s.binpatch", executableBinaryName(), patchFromVersion)
}

// GetStepCompressedExecutablePathForVersion stores the binary patch to restore the precompiled binaries of the step from the latest version
func GetStepCompressedExecutablePathForVersion(fromPatchVersion string, route SteplibRoute, id, version string) string {
	return filepath.Join(GetStepBinDirPath(route, id), version, binaryPatchName(fromPatchVersion))
}

// GetStepExecutablePathForVersion stores the uncompressed precompiled binaries of the step
func GetStepExecutablePathForVersion(route SteplibRoute, id, version string) string {
	return filepath.Join(GetStepBinDirPath(route, id), version, executableBinaryName())
}

// GetStepExecutableChecksumPathForVersion stores the checksum of the uncompressed precompiled binaries of the step
func GetStepExecutableChecksumPathForVersion(route SteplibRoute, id, version string) string {
	return filepath.Join(GetStepBinDirPath(route, id), version, executableChecksumName())
}

// GetStepGlobalInfoPath ...
func GetStepGlobalInfoPath(route SteplibRoute, id string) string {
	return filepath.Join(GetLibraryBaseDirPath(route), "steps", id, "step-info.yml")
}

// GetStepCollectionDirPath ...
// Step's Collection dir path, where it's spec (step.yml) lives.
func GetStepCollectionDirPath(route SteplibRoute, id, version string) string {
	return filepath.Join(GetLibraryBaseDirPath(route), "steps", id, version)
}

// GetStepmanDirPath ...
func GetStepmanDirPath() string {
	return filepath.Join(pathutil.UserHomeDir(), StepmanDirname)
}

// GetCollectionsDirPath ...
func GetCollectionsDirPath() string {
	return filepath.Join(GetStepmanDirPath(), CollectionsDirname)
}

func getRoutingFilePath() string {
	return filepath.Join(GetStepmanDirPath(), RoutingFilename)
}
