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
func New(log stepman.Logger, inventoryURL string) *Client {
	return &Client{
		log:          log,
		inventoryURL: inventoryURL,
		api:          NewHTTPAPI(inventoryURL, httpfetch.NewClient(log)),
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
