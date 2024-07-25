package configmerge

import (
	"io"
	"os"
	"path/filepath"

	pathutilV2 "github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type RepoCache interface {
	GetRepo(ref ConfigReference) string
	SetRepo(dir string, ref ConfigReference)
}

type fileReader struct {
	repoCache RepoCache
	tmpDir    string
	logger    Logger
}

func NewConfigReader(repoCache RepoCache, logger Logger) (ConfigReader, error) {
	tmpDir, err := pathutilV2.NewPathProvider().CreateTempDir("config-merge")
	if err != nil {
		return nil, err
	}

	return fileReader{
		repoCache: repoCache,
		tmpDir:    tmpDir,
		logger:    logger,
	}, nil
}

func (f fileReader) Read(ref ConfigReference) ([]byte, error) {
	if isLocalReference(ref) {
		return f.readFileFromFileSystem(ref.Path)
	}

	cachedRepoDir := f.repoCache.GetRepo(ref)
	if cachedRepoDir != "" {
		pth := filepath.Join(cachedRepoDir, ref.Path)
		return f.readFileFromFileSystem(pth)
	}

	repoDir, err := f.cloneGitRepository(ref)
	if err != nil {
		return nil, err
	}

	f.repoCache.SetRepo(repoDir, ref)
	pth := filepath.Join(repoDir, ref.Path)
	return f.readFileFromFileSystem(pth)
}

func (f fileReader) CleanupRepoDirs() error {
	return os.RemoveAll(f.tmpDir)
}

func isLocalReference(reference ConfigReference) bool {
	return reference.Repository == ""
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

func (f fileReader) cloneGitRepository(ref ConfigReference) (string, error) {
	opts := git.CloneOptions{
		URL: ref.Repository,
	}
	if ref.Branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(ref.Branch)
	}

	repoDir := filepath.Join(f.tmpDir, ref.RepoKey())

	repo, cloneErr := git.PlainClone(repoDir, false, &opts)
	if cloneErr != nil {
		if !isHttpFormatRepoURL(ref.Repository) {
			return "", cloneErr
		}

		// Try repo url with ssh syntax
		repoURL, err := parseGitRepoURL(ref.Repository)
		if err != nil {
			return "", err
		}
		if repoURL.User == "" {
			repoURL.User = "git"
		}

		opts.URL = generateSCPStyleSSHFormatRepoURL(repoURL)
		repo, err = git.PlainClone(repoDir, false, &opts)
		if err != nil {
			// Return the original error
			return "", cloneErr
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	if ref.Commit != "" {
		h, err := repo.ResolveRevision(plumbing.Revision(ref.Commit))
		if err != nil {
			return "", err
		}

		if err := tree.Checkout(&git.CheckoutOptions{
			Hash: *h,
		}); err != nil {
			return "", err
		}
	} else if ref.Tag != "" {
		if err := tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(ref.Tag),
		}); err != nil {
			return "", err
		}
	}

	return repoDir, nil
}
