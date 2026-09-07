package steplibrary

import (
	"context"
	"fmt"

	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
)

type Client struct {
	log          stepman.Logger
	inventoryURL string
	api          API
}

// New builds a stepman.Client.
// inventoryURL: the base URL of the API where metadata is fetched from.
// fetcher: the HTTP client used for every inventory request. Callers pass one
// in so that a single client, and therefore a single connection pool, can be
// shared across everything a run fetches.
func New(log stepman.Logger, inventoryURL string, fetcher httpfetch.Client) *Client {
	return &Client{
		log:          log,
		inventoryURL: inventoryURL,
		api:          NewHTTPAPI(inventoryURL, fetcher),
	}
}

func (c *Client) FetchStepMetadata(ctx context.Context, stepRef stepid.CanonicalID) (models.StepInfoModel, error) {
	stepInfo, resolved, err := c.getStepVersionInfo(ctx, stepRef.IDorURI, stepRef.Version)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("resolve step version: %w", err)
	}

	stepModel, err := c.api.GetStepModel(ctx, resolved)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("fetch step definition: %w", err)
	}
	stepInfo.Step = stepModel

	return stepInfo, nil
}

// StepSourceDownloadLocations returns a priority order of step source zip download locations
func (c *Client) StepSourceDownloadLocations(ctx context.Context, id, version, sourceGit string) ([]models.DownloadLocationModel, error) {
	bases, err := c.api.GetStepSourceDownloadLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch download locations: %w", err)
	}
	return models.BuildStepSourceDownloadLocations(bases, id, version, sourceGit)
}
