package steplibrary

import (
	"context"
	"fmt"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/stepman"
	"gopkg.in/yaml.v2"
)

type Steplib struct {
	log stepman.Logger
	// steplibURI is set by the `default_step_lib_source` property in bitrise.yml
	steplibURI  string
	api         API
	fileManager fileutil.FileManager
}

type ActivateOutputPaths struct {
	YMLPath, CodePath string
}

// New builds a Steplib. steplibURI is the steplib identity; inventoryURL is
// the base URL the V2 inventory JSON is fetched from.
func New(log stepman.Logger, steplibURI, inventoryURL string, fileManager fileutil.FileManager) *Steplib {
	return &Steplib{
		log:         log,
		steplibURI:  steplibURI,
		api:         NewHTTPAPI(inventoryURL, httpfetch.NewClient(log)),
		fileManager: fileManager,
	}
}

func (s *Steplib) Activate(ctx context.Context, stepID, version string, outputPaths ActivateOutputPaths) (ActivatedStep, error) {
	stepInfo, resolved, err := s.getStepVersionInfo(ctx, stepID, version)
	if err != nil {
		return ActivatedStep{}, fmt.Errorf("resolve step version: %w", err)
	}

	stepModel, err := s.api.GetStepModel(ctx, resolved)
	if err != nil {
		return ActivatedStep{}, fmt.Errorf("fetch step definition: %w", err)
	}

	stepYML, err := yaml.Marshal(stepModel)
	if err != nil {
		return ActivatedStep{}, fmt.Errorf("marshal step model to YAML: %w", err)
	}

	if err := s.fileManager.WriteBytes(outputPaths.YMLPath, stepYML); err != nil {
		return ActivatedStep{}, fmt.Errorf("write step.yml: %w", err)
	}

	return ActivatedStep{
		StepInfo:         stepInfo,
		StepYMLPath:      outputPaths.YMLPath,
		ExecutablePath:   "",
		ActivationType:   ActivationTypeSteplibSource,
		DidStepLibUpdate: false, // deprecated
	}, nil
}
