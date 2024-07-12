package configmerge

import (
	"io"
	"os"
	"path/filepath"

	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

type fileReader struct {
	logger logV2.Logger
}

func NewConfigReader(logger logV2.Logger) ConfigReader {
	return fileReader{
		logger: logger,
	}
}

func (f fileReader) Read(ref ConfigReference, dir string) ([]byte, error) {
	if isLocalReference(ref) {
		pth := ref.Path
		if !filepath.IsAbs(pth) {
			pth = filepath.Join(dir, pth)
		}
		return f.readFileFromFileSystem(pth)
	}
	return f.readFileFromGitRepository(ref.Repository, ref.Branch, ref.Commit, ref.Tag, ref.Path)
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
