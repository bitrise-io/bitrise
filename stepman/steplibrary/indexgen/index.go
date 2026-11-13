package indexgen

import (
	"fmt"
	"io/fs"
	"strconv"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
)

// writeStepFiles emits the per-step source files under steps/<id>/. Output paths
// come from steplibindex; the asset src is a source-steplib (input) path.
func writeStepFiles(w *writer, inputFS fs.FS, s parsedStep) error {
	infoPath, err := steplibindex.StepInfoPath(s.id)
	if err != nil {
		return err
	}
	if err := w.writeJSON(infoPath.FS(), s.info); err != nil {
		return err
	}
	for _, f := range s.assetFiles {
		src := "steps/" + s.id + "/assets/" + f
		dst, err := steplibindex.StepAssetPath(s.id, f)
		if err != nil {
			return err
		}
		if err := w.copyFileFromFS(inputFS, src, dst.FS()); err != nil {
			return fmt.Errorf("copy asset %s: %w", src, err)
		}
	}
	for _, v := range s.versions {
		stepJSONPath, err := steplibindex.StepJSONPath(s.id, v.version)
		if err != nil {
			return err
		}
		if err := w.writeJSON(stepJSONPath.FS(), v.model); err != nil {
			return err
		}
	}
	return nil
}

// writeIndexFiles emits the derived index files under index/.
func writeIndexFiles(w *writer, steps []parsedStep) error {
	ids := make([]string, len(steps))
	for i, s := range steps {
		ids[i] = s.id
	}
	if err := w.writeJSON(steplibindex.StepIDsPath().FS(), steplibindex.StepIDs{StepIDs: ids}); err != nil {
		return err
	}

	for _, s := range steps {
		latestPath, err := steplibindex.LatestPointerPath(s.id)
		if err != nil {
			return err
		}
		if err := w.writeJSON(latestPath.FS(), buildLatestPointer(s)); err != nil {
			return err
		}
		versionsPath, err := steplibindex.VersionsPath(s.id)
		if err != nil {
			return err
		}
		if err := w.writeJSON(versionsPath.FS(), buildVersionsJSON(s)); err != nil {
			return err
		}
	}
	return nil
}

func buildLatestPointer(s parsedStep) steplibindex.LatestPointer {
	byMajor := map[string]models.Semver{}
	for _, v := range s.versions {
		majorKey := strconv.FormatUint(v.semver.Major, 10)
		cur, ok := byMajor[majorKey]
		if !ok || models.CmpSemver(v.semver, cur) > 0 {
			byMajor[majorKey] = v.semver
		}
	}
	latestByMajor := make(map[string]string, len(byMajor))
	for k, v := range byMajor {
		latestByMajor[k] = v.String()
	}
	return steplibindex.LatestPointer{
		StepID:        s.id,
		Latest:        s.latest().version,
		LatestByMajor: latestByMajor,
	}
}

func buildVersionsJSON(s parsedStep) steplibindex.Versions {
	versions := make([]string, 0, len(s.versions))
	// Newest-first order: walk the ascending-sorted versions in reverse.
	for i := len(s.versions) - 1; i >= 0; i-- {
		versions = append(versions, s.versions[i].version)
	}
	return steplibindex.Versions{
		StepID:   s.id,
		Versions: versions,
	}
}
