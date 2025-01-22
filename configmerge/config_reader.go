package configmerge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pathutilV2 "github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type fileReader struct {
	logger    Logger
	tmpDir    string
	repoCache map[string]string
	repoURL   GitRepoURL
}

func NewConfigReader(logger Logger) (ConfigReader, error) {
	tmpDir, err := pathutilV2.NewPathProvider().CreateTempDir("config-merge")
	if err != nil {
		return nil, err
	}

	repoURL, err := getRepositoryURL()
	if err != nil {
		return nil, err
	}

	return fileReader{
		logger:    logger,
		tmpDir:    tmpDir,
		repoCache: map[string]string{},
		repoURL:   *repoURL,
	}, nil
}

func (f fileReader) Read(ref ConfigReference) ([]byte, error) {
	if isLocalReference(ref) {
		return f.readFileFromFileSystem(ref.Path)
	}

	cachedRepoDir := f.getRepo(ref)
	if cachedRepoDir != "" {
		pth := filepath.Join(cachedRepoDir, ref.Path)
		return f.readFileFromFileSystem(pth)
	}

	repoURL, err := f.createRepoURL(ref.Repository)
	if err != nil {
		return nil, err
	}

	repoDir := filepath.Join(f.tmpDir, ref.RepoKey())
	if err := f.cloneGitRepository(repoDir, repoURL, ref.Branch, ref.Tag, ref.Commit); err != nil {
		return nil, err
	}

	f.setRepo(repoDir, ref)

	pth := filepath.Join(repoDir, ref.Path)
	return f.readFileFromFileSystem(pth)
}

func (f fileReader) CleanupRepoDirs() error {
	return os.RemoveAll(f.tmpDir)
}

func (f fileReader) readFileFromFileSystem(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			f.logger.Warnf("Failed to close file: %s", err)
		}
	}()
	return io.ReadAll(file)
}

func (f fileReader) createRepoURL(repoName string) (string, error) {
	repoURL := f.repoURL
	pathComponents := strings.Split(repoURL.Path, "/")
	if len(pathComponents) < 2 {
		return "", fmt.Errorf("invalid repository path: %s", repoURL.Path)
	}
	repoURL.Path = strings.Join(pathComponents[:len(pathComponents)-2], "/") + "/" + repoName + ".git"

	return repoURL.URLString(repoURL.OriginalSyntax), nil
}

func (f fileReader) cloneGitRepository(repoDir, repoURL, branch, tag, commit string) error {
	opts := git.CloneOptions{
		URL: repoURL,
	}
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	repo, cloneErr := git.PlainClone(repoDir, false, &opts)
	if cloneErr != nil {
		// TODO: revisit error handling

		// Try repo url with ssh syntax
		gitRepoURL, err := parseGitRepoURL(repoURL)
		if err != nil {
			return cloneErr
		}
		if gitRepoURL.OriginalSyntax != HTTPSRepoURLSyntax {
			return cloneErr
		}

		if gitRepoURL.User == "" {
			gitRepoURL.User = "git"
		}

		opts.URL = gitRepoURL.URLString(SSHGitRepoURLSyntax)
		repo, err = git.PlainClone(repoDir, false, &opts)
		if err != nil {
			// Return the original error
			return cloneErr
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if commit != "" {
		h, err := repo.ResolveRevision(plumbing.Revision(commit))
		if err != nil {
			return err
		}

		if err := tree.Checkout(&git.CheckoutOptions{
			Hash: *h,
		}); err != nil {
			return err
		}
	} else if tag != "" {
		if err := tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(tag),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (f fileReader) getRepo(ref ConfigReference) string {
	return f.repoCache[ref.RepoKey()]
}

func (f fileReader) setRepo(dir string, ref ConfigReference) {
	f.repoCache[ref.RepoKey()] = dir
}

// TODO: review error messages
func getRepositoryURL() (*GitRepoURL, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	if len(remotes) == 0 {
		return nil, fmt.Errorf("no remotes found")
	}

	var remoteConfig *config.RemoteConfig
	if len(remotes) > 1 {
		for _, remote := range remotes {
			c := remote.Config()
			if c == nil {
				continue
			}

			if c.Name == "origin" {
				remoteConfig = c
			}
		}
	} else if len(remotes) == 1 {
		defaultRemote := remotes[0]
		c := defaultRemote.Config()
		if c == nil {
			return nil, fmt.Errorf("no remote config found")
		}
		remoteConfig = c
	}

	if remoteConfig == nil {
		return nil, fmt.Errorf("no default remote config found")
	}

	if len(remoteConfig.URLs) == 0 {
		return nil, fmt.Errorf("no remote URLs found")
	} else if len(remoteConfig.URLs) > 1 {
		return nil, fmt.Errorf("multiple remote URLs found")
	}

	return parseGitRepoURL(remoteConfig.URLs[0])
}

func isLocalReference(reference ConfigReference) bool {
	return reference.Repository == ""
}
