package configmerge

import (
	"io"
	"os"
	"path/filepath"

	logV2 "github.com/bitrise-io/go-utils/v2/log"
	pathutilV2 "github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

type repoCache struct {
	cache map[string]string
}

type RepoCache interface {
	GetRepo(ref ConfigReference) string
	SetRepo(dir string, ref ConfigReference)
}

func NewRepoCache() RepoCache {
	return repoCache{
		cache: map[string]string{},
	}
}

func (c repoCache) GetRepo(ref ConfigReference) string {
	return c.cache[ref.RepoKey()]
}

func (c repoCache) SetRepo(dir string, ref ConfigReference) {
	c.cache[ref.RepoKey()] = dir
}

type fileReader struct {
	repoCache RepoCache
	logger    logV2.Logger
}

func NewConfigReader(repoCache RepoCache, logger logV2.Logger) ConfigReader {
	return fileReader{
		repoCache: repoCache,
		logger:    logger,
	}
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

	repoDir, err := f.cloneGitRepository(ref.Repository, ref.Branch, ref.Commit, ref.Tag)
	if err != nil {
		return nil, err
	}

	f.repoCache.SetRepo(repoDir, ref)
	pth := filepath.Join(repoDir, ref.Path)
	return f.readFileFromFileSystem(pth)
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

func (f fileReader) cloneGitRepository(repository string, branch string, commit string, tag string) (string, error) {
	opts := git.CloneOptions{
		URL: repository,
	}
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	tmpDir, err := pathutilV2.NewPathProvider().CreateTempDir("config-merge")
	if err != nil {
		return "", err
	}

	repo, cloneErr := git.PlainClone(tmpDir, false, &opts)
	if cloneErr != nil {
		if !isHttpFormatRepoURL(repository) {
			return "", cloneErr
		}

		// Try repo url with ssh syntax
		repoURL, err := parseGitRepoURL(repository)
		if err != nil {
			return "", err
		}
		if repoURL.User == "" {
			repoURL.User = "git"
		}

		opts.URL = generateSCPStyleSSHFormatRepoURL(repoURL)
		repo, err = git.PlainClone(tmpDir, false, &opts)
		if err != nil {
			// Return the original error
			return "", cloneErr
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	if commit != "" {
		h, err := repo.ResolveRevision(plumbing.Revision(commit))
		if err != nil {
			return "", err
		}

		if err := tree.Checkout(&git.CheckoutOptions{
			Hash: *h,
		}); err != nil {
			return "", err
		}
	} else if tag != "" {
		if err := tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(tag),
		}); err != nil {
			return "", err
		}
	}

	return tmpDir, nil
}

func (f fileReader) readFileFromGitRepository(repository string, branch string, commit string, tag string, path string) ([]byte, error) {
	opts := git.CloneOptions{
		URL: repository,
	}
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	repo, cloneErr := git.Clone(memory.NewStorage(), memfs.New(), &opts)
	if cloneErr != nil {
		if !isHttpFormatRepoURL(repository) {
			return nil, cloneErr
		}

		// Try repo url with ssh syntax
		repoURL, err := parseGitRepoURL(repository)
		if err != nil {
			return nil, err
		}
		if repoURL.User == "" {
			repoURL.User = "git"
		}

		opts := git.CloneOptions{
			URL: generateSCPStyleSSHFormatRepoURL(repoURL),
		}
		if branch != "" {
			opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		}

		repo, err = git.Clone(memory.NewStorage(), memfs.New(), &opts)
		if err != nil {
			// Return the original error
			return nil, cloneErr
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	if commit != "" {
		h, err := repo.ResolveRevision(plumbing.Revision(commit))
		if err != nil {
			return nil, err
		}

		if err := tree.Checkout(&git.CheckoutOptions{
			Hash: *h,
		}); err != nil {
			return nil, err
		}
	} else if tag != "" {
		if err := tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(tag),
		}); err != nil {
			return nil, err
		}
	}

	file, err := tree.Filesystem.Open(path)
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
