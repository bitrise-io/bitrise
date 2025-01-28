package configmerge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/log"
	pathutilV2 "github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	currentRepositoryURLEnvKey = "BITRISE_CURRENT_REPOSITORY_URL"
)

type fileReader struct {
	logger    log.Logger
	tmpDir    string
	repoCache map[string]string
	repoURL   *GitRepoURL
}

func NewConfigReader(logger log.Logger) (ConfigReader, error) {
	tmpDir, err := pathutilV2.NewPathProvider().CreateTempDir("config-merge")
	if err != nil {
		return nil, err
	}

	return &fileReader{
		logger:    logger,
		tmpDir:    tmpDir,
		repoCache: map[string]string{},
	}, nil
}

func (f *fileReader) Read(ref ConfigReference) ([]byte, error) {
	if ref.IsLocalReference() {
		f.logger.Debugf("reading local config module at: %s", ref.Path)
		return f.readFileFromFileSystem(ref.Path)
	}

	repoStateText := fmt.Sprintf("on branch '%s'", ref.Branch)
	if ref.Tag != "" {
		repoStateText = fmt.Sprintf("on tag '%s'", ref.Tag)
	} else if ref.Commit != "" {
		repoStateText = fmt.Sprintf("on commit '%s'", ref.Commit)
	}
	f.logger.Debugf("reading remote config module '%s' from repo '%s' %s", ref.Path, ref.Repository, repoStateText)

	cachedRepoDir := f.getRepo(ref)
	if cachedRepoDir != "" {
		pth := filepath.Join(cachedRepoDir, ref.Path)
		f.logger.Debugf("reading config module (%s) from a cached repository: %s", ref.Path, pth)
		return f.readFileFromFileSystem(pth)
	}

	if f.repoURL == nil {
		f.logger.Debugf("getting current repository url")
		if err := f.getCurrentRepositoryURL(); err != nil {
			return nil, fmt.Errorf("failed to get current repository URL: %w, the repository URL can be set manually using the 'BITRISE_CURRENT_REPOSITORY_URL' environment variable", err)
		}
		f.logger.Debugf("current repository url: %s", f.repoURL.URLString(f.repoURL.OriginalSyntax))
	}

	cloneRepoStateText := fmt.Sprintf("with branch '%s'", ref.Branch)
	if ref.Tag != "" {
		cloneRepoStateText = fmt.Sprintf("with tag '%s'", ref.Tag)
	} else if ref.Commit != "" {
		cloneRepoStateText = fmt.Sprintf("with commit '%s'", ref.Commit)
	}

	moduleGitRepoURL := f.repoURL.RepoURLForRepo(ref.Repository)
	moduleRepoURL := moduleGitRepoURL.URLString(moduleGitRepoURL.OriginalSyntax)
	f.logger.Debugf("cloning repository '%s' %s", moduleRepoURL, cloneRepoStateText)
	repoDir := filepath.Join(f.tmpDir, ref.RepoKey())
	if err := f.cloneGitRepository(repoDir, moduleRepoURL, ref.Branch, ref.Tag, ref.Commit); err != nil {
		return nil, err
	}

	f.setRepo(repoDir, ref)

	pth := filepath.Join(repoDir, ref.Path)
	f.logger.Debugf("reading config module (%s) from a cloned repository: %s", ref.Path, pth)
	return f.readFileFromFileSystem(pth)
}

func (f *fileReader) CleanupRepoDirs() error {
	f.logger.Debugf("Cleaning up modular config local cache dir: %s", f.tmpDir)
	return os.RemoveAll(f.tmpDir)
}

func (f *fileReader) readFileFromFileSystem(name string) ([]byte, error) {
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

func (f *fileReader) cloneGitRepository(repoDir, repoURL, branch, tag, commit string) error {
	opts := git.CloneOptions{
		URL: repoURL,
	}
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	repo, cloneErr := git.PlainClone(repoDir, false, &opts)
	if cloneErr != nil {
		f.logger.Warnf("Failed to clone config module repository (%s): %s, trying with a different repository URL syntax...", repoURL, cloneErr)

		// Try repo url with a different syntax
		gitRepoURL, err := NewGitRepoURL(repoURL)
		if err != nil {
			return fmt.Errorf("failed to parse repository URL (%s):  %w", repoURL, err)
		}

		var repoURLSyntax GitRepoURLSyntax
		if gitRepoURL.OriginalSyntax == HTTPSRepoURLSyntax {
			repoURLSyntax = SSHGitRepoURLSyntax
		} else {
			repoURLSyntax = HTTPSRepoURLSyntax
		}

		if gitRepoURL.User == "" {
			gitRepoURL.User = "git"
		}

		opts.URL = gitRepoURL.URLString(repoURLSyntax)
		repo, err = git.PlainClone(repoDir, false, &opts)
		if err != nil {
			return fmt.Errorf("failed to clone repository (%s): %w", opts.URL, err)
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree for repository (%s): %w", opts.URL, err)
	}

	if commit != "" {
		h, err := repo.ResolveRevision(plumbing.Revision(commit))
		if err != nil {
			return fmt.Errorf("failed to resolve commit (%s): %w", commit, err)
		}

		if err := tree.Checkout(&git.CheckoutOptions{
			Hash: *h,
		}); err != nil {
			return fmt.Errorf("failed to checkout commit (%s): %w", commit, err)
		}
	} else if tag != "" {
		if err := tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(tag),
		}); err != nil {
			return fmt.Errorf("failed to checkout tag (%s): %w", tag, err)
		}
	}

	return nil
}

func (f *fileReader) getCurrentRepositoryURL() error {
	if repoURL := os.Getenv(currentRepositoryURLEnvKey); repoURL != "" {
		gitRepoURL, err := NewGitRepoURL(repoURL)
		if err != nil {
			return fmt.Errorf("failed to parse repository URL: %w, the URL is expected in a HTTPS (https://<host>[:<port>]/<path-to-git-repo>) or SSH ([<user>@]<host>:<path-to-git-repo>) syntax ", err)
		}

		f.repoURL = gitRepoURL
		return nil
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not open repository in the working directory: %w", err)
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return fmt.Errorf("could not get remotes for the repository in the working directory: %w", err)
	}
	if len(remotes) == 0 {
		return fmt.Errorf("no remotes found for the repository in the working directory")
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
			return fmt.Errorf("no remote config found for the repository in the working directory")
		}
		remoteConfig = c
	}

	if remoteConfig == nil {
		return fmt.Errorf("no default remote config found for the repository in the working directory")
	}

	if len(remoteConfig.URLs) == 0 {
		return fmt.Errorf("no remote URLs found for the repository in the working directory")
	} else if len(remoteConfig.URLs) > 1 {
		return fmt.Errorf("multiple remote URLs found for the repository in the working directory")
	}

	gitRepoURL, err := NewGitRepoURL(remoteConfig.URLs[0])
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w, the URL is expected in a HTTPS (https://<host>[:<port>]/<path-to-git-repo>) or SSH ([<user>@]<host>:<path-to-git-repo>) syntax ", err)
	}

	f.repoURL = gitRepoURL
	return nil
}

func (f *fileReader) getRepo(ref ConfigReference) string {
	return f.repoCache[ref.RepoKey()]
}

func (f *fileReader) setRepo(dir string, ref ConfigReference) {
	f.repoCache[ref.RepoKey()] = dir
}
