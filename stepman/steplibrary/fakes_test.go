package steplibrary

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
)

// FakeAPI is an in-memory API implementation used as a base for test fakes
// that embed it and selectively override methods.
type FakeAPI struct{}

func (m FakeAPI) GetAllStepIDs(_ context.Context) ([]string, error) {
	return []string{"xcode-test", "script"}, nil
}

func (m FakeAPI) GetLatestStepVersions(_ context.Context, id string) (steplibindex.LatestPointer, error) {
	versions := map[string]steplibindex.LatestPointer{
		"script": {
			StepID: "script",
			Latest: "3.0.0",
			LatestByMajor: map[string]string{
				"1": "1.2.0",
				"2": "2.4.1",
				"3": "3.0.0",
			},
		},
	}

	v, ok := versions[id]
	if !ok {
		return steplibindex.LatestPointer{}, errors.New("not found")
	}
	return v, nil
}

func (m FakeAPI) GetAllStepVersions(_ context.Context, id string) ([]string, error) {
	versions := map[string][]string{
		"script": {"1.0.0", "1.1.5", "1.2.0", "2.0.0", "2.4.0", "2.4.1", "3.0.0"},
	}
	v, ok := versions[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (m FakeAPI) GetStepGroupInfo(_ context.Context, id string) (steplibindex.StepInfo, error) {
	infos := map[string]steplibindex.StepInfo{
		"script": {
			Maintainer:  "bitrise",
			Deprecation: nil,
			AssetURLs:   []string{"assets/icon.svg"},
		},
	}
	v, ok := infos[id]
	if !ok {
		return steplibindex.StepInfo{}, errors.New("not found")
	}
	return v, nil
}

func (m FakeAPI) GetStepModel(_ context.Context, step ResolvedStepVersion) (models.StepModel, error) {
	if step.ID != "script" {
		return models.StepModel{}, errors.New("not found")
	}
	//nolint:exhaustruct // mock returns a minimal StepModel
	return models.StepModel{
		Title:   pointers.NewStringPtr("Script"),
		Summary: pointers.NewStringPtr("Runs a shell script."),
	}, nil
}

// fakeGetFetcher implements httpfetch.Client.Get, returning a fixed body whose
// Close returns closeErr. Used to exercise fetchJSON's close-error path.
type fakeGetFetcher struct {
	body     string
	closeErr error
}

func (f fakeGetFetcher) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return errReadCloser{Reader: strings.NewReader(f.body), closeErr: f.closeErr}, nil
}

func (f fakeGetFetcher) Download(_ context.Context, _, _ string) error {
	return errors.New("Download not used")
}

func (f fakeGetFetcher) DownloadWithHash(_ context.Context, _, _, _ string) error {
	return errors.New("DownloadWithHash not used")
}

// errReadCloser wraps a reader and returns closeErr from Close.
type errReadCloser struct {
	io.Reader
	closeErr error
}

func (e errReadCloser) Close() error { return e.closeErr }
