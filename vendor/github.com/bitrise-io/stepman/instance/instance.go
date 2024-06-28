package instance

import (
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

type Instance struct {
	stepCollectionCache map[string]models.StepCollectionModel
	logger              stepman.Logger
}

func NewInstance(logger stepman.Logger) Instance {
	return Instance{
		stepCollectionCache: make(map[string]models.StepCollectionModel),
		logger:              logger,
	}
}

func (i Instance) Update(uri string) error {
	if _, ok := i.stepCollectionCache[uri]; ok {
		return nil
	}

	stepCollection, err := stepman.UpdateLibrary(uri, i.logger)
	if err != nil {
		return err
	}

	i.stepCollectionCache[uri] = stepCollection

	return nil
}

func (i Instance) QueryStepInfo(uri, id, version string) (models.StepInfoModel, error) {
	switch uri {
	case "git":
		return stepman.QueryStepInfoFromGit(id, version)
	case "path":
		return stepman.QueryStepInfoFromPath(id)
	}

	stepCollection, err := i.getCollection(uri)
	if err != nil {
		return models.StepInfoModel{}, err
	}

	return stepman.QueryStepInfoFromCollection(stepCollection, uri, id, version)
}

func (i Instance) Activate(uri, id, version, dir, ymlPth string, isOfflineMode bool) error {
	stepCollection, err := i.getCollection(uri)
	if err != nil {
		return err
	}

	return stepman.ActivateFromCollection(stepCollection, uri, id, version, dir, ymlPth, false, i.logger, isOfflineMode)
}

func (i Instance) getCollection(uri string) (models.StepCollectionModel, error) {
	stepCollection, ok := i.stepCollectionCache[uri]
	if ok {
		return stepCollection, nil
	}

	if err := i.Update(uri); err != nil {
		return models.StepCollectionModel{}, err
	}

	return i.stepCollectionCache[uri], nil
}
