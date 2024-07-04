package configmerge

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v2"
)

type repoInfo struct {
	defaultRemoteURL string
	branch           string
	commit           string
	tag              string
}

type Merger struct {
	repo     *git.Repository
	repoInfo repoInfo
}

func NewMerger(repoPth string) (*Merger, error) {
	repo, err := git.PlainOpen(repoPth)
	if err != nil {
		return nil, err
	}

	info, err := readRepoInfo(repo)
	if err != nil {
		return nil, err
	}

	return &Merger{
		repo:     repo,
		repoInfo: *info,
	}, nil
}

func (m Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	f, err := os.Open(mainConfigPth)
	if err != nil {
		return "", nil, err
	}

	ref := configReference{
		Repository: m.repoInfo.defaultRemoteURL,
		Commit:     m.repoInfo.commit,
		Tag:        m.repoInfo.tag,
		Branch:     m.repoInfo.branch,
		Path:       mainConfigPth,
	}

	configTree, err := m.buildConfigTree(f, ref)
	if err != nil {
		return "", nil, err
	}

	mergedConfigContent, err := configTree.Merge()
	if err != nil {
		return "", nil, err
	}

	return mergedConfigContent, configTree, nil
}

func (m Merger) buildConfigTree(configReader io.Reader, reference configReference) (*models.ConfigFileTreeModel, error) {
	b, err := io.ReadAll(configReader)
	if err != nil {
		return nil, err
	}

	var config struct {
		Include []configReference `yaml:"include" json:"include"`
	}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	var includedConfigs []models.ConfigFileTreeModel
	for _, include := range config.Include {
		reader, err := openConfigModule(include, m.repoInfo)
		if err != nil {
			return nil, err
		}

		tree, err := m.buildConfigTree(reader, include)
		if err != nil {
			return nil, err
		}

		includedConfigs = append(includedConfigs, *tree)
	}

	return &models.ConfigFileTreeModel{
		Path:     reference.Key(),
		Contents: string(b),
		Includes: includedConfigs,
	}, nil
}

func openConfigModule(reference configReference, info repoInfo) (io.Reader, error) {
	if isLocalReference(reference, info) {
		return openLocalConfigModule(reference)
	} else {
		return openRemoteConfigModule(reference)
	}
}

func isLocalReference(reference configReference, info repoInfo) bool {
	if reference.Repository == "" {
		return true
	}

	if getRepo(reference.Repository) != getRepo(info.defaultRemoteURL) {
		return false
	}

	switch {
	case reference.Commit != "":
		return reference.Commit == info.commit || reference.Commit == info.commit[:7]
	case reference.Tag != "":
		return reference.Tag == info.tag
	case reference.Branch != "":
		return reference.Branch == info.branch
	}

	return false
}

func openLocalConfigModule(reference configReference) (io.Reader, error) {
	return os.Open(reference.Path)
}

func openRemoteConfigModule(reference configReference) (io.Reader, error) {
	return nil, nil
}

func readRepoInfo(repo *git.Repository) (*repoInfo, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	// Get branch name
	var branch string
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	}

	// Get commit hash
	commitObj, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}
	commit := commitObj.Hash.String()

	// Get tag name
	var tag string
	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		obj, err := repo.TagObject(ref.Hash())
		if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
			return err
		}

		if obj.Target == head.Hash() {
			tag = ref.Name().Short()
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Get remote URL
	config, err := repo.Config()
	if err != nil {
		return nil, err
	}

	remote, ok := config.Remotes[git.DefaultRemoteName]
	if !ok {
		return nil, fmt.Errorf("remote %s not exists", git.DefaultRemoteName)
	}

	if len(remote.URLs) == 0 {
		return nil, fmt.Errorf("URL not set for remote %s", git.DefaultRemoteName)
	}

	remoteURL := remote.URLs[0]

	return &repoInfo{
		defaultRemoteURL: remoteURL,
		branch:           branch,
		commit:           commit,
		tag:              tag,
	}, nil
}

func getRepo(url string) string {
	var host, repo string
	switch {
	case strings.HasPrefix(url, "https://"):
		url = strings.TrimPrefix(url, "https://")
		idx := strings.Index(url, "/")
		host, repo = url[:idx], url[idx+1:]
	case strings.HasPrefix(url, "git@"):
		url = url[strings.Index(url, "@")+1:]
		idx := strings.Index(url, ":")
		host, repo = url[:idx], url[idx+1:]
	case strings.HasPrefix(url, "ssh://"):
		url = url[strings.Index(url, "@")+1:]
		if strings.Contains(url, ":") {
			idxColon, idxSlash := strings.Index(url, ":"), strings.Index(url, "/")
			host, repo = url[:idxColon], url[idxSlash+1:]
		} else {
			idx := strings.Index(url, "/")
			host, repo = url[:idx], url[idx+1:]
		}
	}
	return host + "/" + strings.TrimSuffix(repo, ".git")
}
