package specgen

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/bitrise-io/stepman/steplibrary/spec"
	"gopkg.in/yaml.v2"
)

// Options control generator behavior. Zero values are filled with sensible
// defaults; callers (CLI / tests) override what they need.
type Options struct {
	// GeneratedAt is written to meta.json and latest_versions.json.
	// Tests should set this for deterministic output.
	GeneratedAt time.Time
	// SteplibCommitSHA, if set, is written to meta.json and latest_versions.json.
	SteplibCommitSHA string
}

// Stats summarizes a successful generation.
type Stats struct {
	StepCount    int
	VersionCount int
	FilesWritten int
	BytesWritten int64
	Duration     time.Duration
}

// Generate sets up the steplib identified by steplibURI (cloning it into
// stepman's local cache via stepman.SetupLibrary if not already present) and
// writes the V2 inventory tree to outputDir. It is the URI-based entry point
// used by the CLI; GenerateFromSteplibClone is the lower-level core that reads
// from an already-available filesystem.
func Generate(steplibURI, outputDir string, opts Options, log stepman.Logger) (Stats, error) {
	if err := stepman.SetupLibrary(steplibURI, log); err != nil {
		return Stats{}, fmt.Errorf("setup steplib %s: %w", steplibURI, err)
	}
	route, found := stepman.ReadRoute(steplibURI)
	if !found {
		return Stats{}, fmt.Errorf("no route for steplib %s after setup", steplibURI)
	}
	return GenerateFromSteplibClone(os.DirFS(stepman.GetLibraryBaseDirPath(route)), outputDir, opts, log)
}

// GenerateFromSteplibClone reads a bitrise-steplib clone from inputFS and writes the V2 inventory
// tree to outputDir. It is destructive in the sense that it writes files; it
// does NOT delete existing files outside the paths it owns.
func GenerateFromSteplibClone(inputFS fs.FS, outputDir string, opts Options, log stepman.Logger) (Stats, error) {
	start := time.Now()
	opts = withDefaults(opts)

	steplibYML, err := readSteplibYML(inputFS)
	if err != nil {
		return Stats{}, fmt.Errorf("read steplib.yml: %w", err)
	}

	steps, err := collectSteps(inputFS, log)
	if err != nil {
		return Stats{}, err
	}

	w := &writer{outputDir: outputDir, fw: realFileWriter{}, fileCount: 0, byteCount: 0}

	for _, s := range steps {
		if err := writeStepFiles(w, inputFS, s); err != nil {
			return Stats{}, fmt.Errorf("write step %s: %w", s.id, err)
		}
	}

	if err := writeSpecFiles(w, steps, opts); err != nil {
		return Stats{}, fmt.Errorf("write spec files: %w", err)
	}

	meta := spec.Meta{
		FormatVersion:     spec.FormatVersion,
		UpdatedAt:         opts.GeneratedAt,
		SteplibCommitSHA:  opts.SteplibCommitSHA,
		SteplibSource:     steplibYML.SteplibSource,
		DownloadLocations: steplibYML.DownloadLocations,
	}
	if err := w.writeJSON("meta.json", meta); err != nil {
		return Stats{}, fmt.Errorf("write meta.json: %w", err)
	}

	versionCount := 0
	for _, s := range steps {
		versionCount += len(s.versions)
	}
	return Stats{
		StepCount:    len(steps),
		VersionCount: versionCount,
		FilesWritten: w.fileCount,
		BytesWritten: w.byteCount,
		Duration:     time.Since(start),
	}, nil
}

// withDefaults fills zero-valued options.
func withDefaults(o Options) Options {
	if o.GeneratedAt.IsZero() {
		o.GeneratedAt = time.Now().UTC()
	}
	return o
}

// parsedStep is the intermediate representation collected during the walk
// phase, used by the write phase to emit per-step and index files.
type parsedStep struct {
	id          string
	info        spec.StepInfo // step-info.yml + assets/ listing
	hasInfoFile bool         // whether step-info.yml existed
	assetFiles  []string     // relative paths under assets/, sorted
	versions    map[string]models.StepModel
	versionList []string // sorted ascending by semver
	latest      string   // highest semver in versionList
}

func collectSteps(inputFS fs.FS, log stepman.Logger) ([]parsedStep, error) {
	entries, err := fs.ReadDir(inputFS, "steps")
	if err != nil {
		return nil, fmt.Errorf("read steps: %w", err)
	}

	var out []parsedStep
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		s, err := collectStep(inputFS, e.Name(), log)
		if err != nil {
			return nil, err
		}
		if len(s.versions) == 0 {
			log.Warnf("step %s has no parseable versions, skipping", s.id)
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out, nil
}

func collectStep(inputFS fs.FS, id string, log stepman.Logger) (parsedStep, error) {
	s := parsedStep{
		id:          id,
		info:        spec.StepInfo{Maintainer: "", Deprecation: nil, AssetURLs: nil},
		hasInfoFile: false,
		assetFiles:  nil,
		versions:    map[string]models.StepModel{},
		versionList: nil,
		latest:      "",
	}
	stepDir := "steps/" + id

	info, hasInfo, err := readStepGroupInfo(inputFS, stepDir+"/step-info.yml")
	if err != nil {
		return s, fmt.Errorf("read step-info.yml for %s: %w", id, err)
	}
	s.info = info
	s.hasInfoFile = hasInfo

	assetFiles, err := listAssets(inputFS, stepDir+"/assets")
	if err != nil {
		return s, fmt.Errorf("list assets for %s: %w", id, err)
	}
	s.assetFiles = assetFiles
	if len(assetFiles) > 0 {
		if s.info.AssetURLs == nil {
			s.info.AssetURLs = make(map[string]string, len(assetFiles))
		}
		for _, f := range assetFiles {
			s.info.AssetURLs[f] = "assets/" + f
		}
	}

	subEntries, err := fs.ReadDir(inputFS, stepDir)
	if err != nil {
		return s, fmt.Errorf("read %s: %w", stepDir, err)
	}
	for _, sub := range subEntries {
		if !sub.IsDir() {
			continue
		}
		if sub.Name() == "assets" {
			continue
		}
		if _, err := models.ParseSemver(sub.Name()); err != nil {
			log.Warnf("step %s: version dir %q is not semver, skipping", id, sub.Name())
			continue
		}
		step, err := parseStepYML(inputFS, stepDir+"/"+sub.Name()+"/step.yml")
		if err != nil {
			return s, fmt.Errorf("parse %s/%s: %w", id, sub.Name(), err)
		}
		s.versions[sub.Name()] = step
	}

	s.versionList = sortedSemver(s.versions)
	if len(s.versionList) > 0 {
		s.latest = s.versionList[len(s.versionList)-1]
	}
	return s, nil
}

func readSteplibYML(inputFS fs.FS) (models.StepCollectionModel, error) {
	bytes, err := fs.ReadFile(inputFS, "steplib.yml")
	if err != nil {
		return models.StepCollectionModel{}, err
	}
	var c models.StepCollectionModel
	if err := yaml.Unmarshal(bytes, &c); err != nil {
		return models.StepCollectionModel{}, err
	}
	return c, nil
}

func readStepGroupInfo(inputFS fs.FS, path string) (spec.StepInfo, bool, error) {
	bytes, err := fs.ReadFile(inputFS, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return spec.StepInfo{}, false, nil
		}
		return spec.StepInfo{}, false, err
	}
	var sgi models.StepGroupInfoModel
	if err := yaml.Unmarshal(bytes, &sgi); err != nil {
		return spec.StepInfo{}, true, err
	}
	out := spec.StepInfo{
		Maintainer:  sgi.Maintainer,
		Deprecation: nil,
		AssetURLs:   nil,
	}
	if sgi.RemovalDate != "" || sgi.DeprecateNotes != "" {
		out.Deprecation = &spec.Deprecation{
			RemovalDate: sgi.RemovalDate,
			Notes:       sgi.DeprecateNotes,
		}
	}
	return out, true, nil
}

func listAssets(inputFS fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(inputFS, dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

func parseStepYML(inputFS fs.FS, path string) (models.StepModel, error) {
	bytes, err := fs.ReadFile(inputFS, path)
	if err != nil {
		return models.StepModel{}, err
	}
	var step models.StepModel
	if err := yaml.Unmarshal(bytes, &step); err != nil {
		return models.StepModel{}, fmt.Errorf("yaml unmarshal: %w", err)
	}
	if err := step.Normalize(); err != nil {
		return models.StepModel{}, fmt.Errorf("normalize: %w", err)
	}
	return step, nil
}

// sortedSemver returns the keys of m sorted ascending by semver. Keys that
// don't parse as semver are silently dropped (collectStep already warned).
func sortedSemver(m map[string]models.StepModel) []string {
	parsed := make([]models.Semver, 0, len(m))
	keyByStr := make(map[string]string, len(m))
	for k := range m {
		v, err := models.ParseSemver(k)
		if err != nil {
			continue
		}
		parsed = append(parsed, v)
		keyByStr[v.String()] = k
	}
	sort.Slice(parsed, func(i, j int) bool { return models.CmpSemver(parsed[i], parsed[j]) < 0 })
	out := make([]string, 0, len(parsed))
	for _, v := range parsed {
		out = append(out, keyByStr[v.String()])
	}
	return out
}

// ---------------------------------------------------------------------------
// step-level writes (steps/<id>/...)
// ---------------------------------------------------------------------------

func writeStepFiles(w *writer, inputFS fs.FS, s parsedStep) error {
	if s.hasInfoFile || len(s.assetFiles) > 0 {
		if err := w.writeJSON(filepath.Join("steps", s.id, "step-info.json"), s.info); err != nil {
			return err
		}
	}
	for _, f := range s.assetFiles {
		src := "steps/" + s.id + "/assets/" + f
		dst := filepath.Join("steps", s.id, "assets", f)
		if err := w.copyFileFromFS(inputFS, src, dst); err != nil {
			return fmt.Errorf("copy asset %s: %w", src, err)
		}
	}
	for _, v := range s.versionList {
		step := s.versions[v]
		if err := w.writeJSON(filepath.Join("steps", s.id, v, "step.json"), step); err != nil {
			return err
		}
	}
	return nil
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---------------------------------------------------------------------------
// spec/ writes (derived index files)
// ---------------------------------------------------------------------------

func writeSpecFiles(w *writer, steps []parsedStep, opts Options) error {
	ids := make([]string, len(steps))
	for i, s := range steps {
		ids[i] = s.id
	}
	if err := w.writeJSON("spec/step_ids.json", spec.StepIDs{StepIDs: ids}); err != nil {
		return err
	}

	catalog := buildCatalog(steps, opts)
	if err := w.writeJSON("spec/latest_versions.json", catalog); err != nil {
		return err
	}

	for _, s := range steps {
		if err := w.writeJSON(filepath.Join("spec", "steps", s.id, "latest.json"), buildLatestPointer(s)); err != nil {
			return err
		}
		if err := w.writeJSON(filepath.Join("spec", "steps", s.id, "versions.json"), buildVersionsJSON(s)); err != nil {
			return err
		}
	}
	return nil
}

func buildCatalog(steps []parsedStep, opts Options) spec.Catalog {
	out := spec.Catalog{
		GeneratedAt:      opts.GeneratedAt,
		SteplibCommitSHA: opts.SteplibCommitSHA,
		Steps:            make(map[string]spec.CatalogEntry, len(steps)),
	}
	for _, s := range steps {
		out.Steps[s.id] = buildCatalogEntry(s)
	}
	return out
}

func buildCatalogEntry(s parsedStep) spec.CatalogEntry {
	latestStep := s.versions[s.latest]

	var publishedAt *time.Time
	if latestStep.PublishedAt != nil && !latestStep.PublishedAt.IsZero() {
		publishedAt = latestStep.PublishedAt
	}

	// Catalog asset URLs are INVENTORY-ROOT-RELATIVE. Catalog consumers
	// resolve them against the inventory base URL (i.e., the URL the
	// catalog itself was fetched from, with /spec/latest_versions.json
	// trimmed). This keeps the V2 inventory portable across hosting
	// changes — no V1-era S3 host is baked into the catalog payload.
	var assetURLs map[string]string
	if len(s.info.AssetURLs) > 0 {
		assetURLs = make(map[string]string, len(s.info.AssetURLs))
		for filename, relPath := range s.info.AssetURLs {
			assetURLs[filename] = catalogAssetURL(s.id, relPath)
		}
	}

	return spec.CatalogEntry{
		LatestVersion:   s.latest,
		PublishedAt:     publishedAt,
		Title:           derefStr(latestStep.Title),
		Summary:         derefStr(latestStep.Summary),
		Maintainer:      s.info.Maintainer,
		TypeTags:        latestStep.TypeTags,
		ProjectTypeTags: latestStep.ProjectTypeTags,
		HostOsTags:      latestStep.HostOsTags,
		Website:         derefStr(latestStep.Website),
		SourceCodeURL:   derefStr(latestStep.SourceCodeURL),
		SupportURL:      derefStr(latestStep.SupportURL),
		AssetURLs:       assetURLs,
		HasExecutable:   latestStep.Executables != nil && len(*latestStep.Executables) > 0,
		Deprecation:     s.info.Deprecation,
	}
}

// catalogAssetURL produces the inventory-root-relative path the catalog
// emits for a given asset. The relPath comes from step-info.json (which
// is step-dir-relative, e.g. "assets/icon.svg"); we prepend "steps/<id>/"
// so the result is anchored at the inventory root.
func catalogAssetURL(stepID, relPath string) string {
	return "steps/" + stepID + "/" + relPath
}

func buildLatestPointer(s parsedStep) spec.LatestPointer {
	byMajor := map[string]models.Semver{}
	for _, v := range s.versionList {
		sv, err := models.ParseSemver(v)
		if err != nil {
			continue
		}
		majorKey := strconv.FormatUint(sv.Major, 10)
		cur, ok := byMajor[majorKey]
		if !ok || models.CmpSemver(sv, cur) > 0 {
			byMajor[majorKey] = sv
		}
	}
	latestByMajor := make(map[string]string, len(byMajor))
	for k, v := range byMajor {
		latestByMajor[k] = v.String()
	}
	return spec.LatestPointer{
		StepID:        s.id,
		Latest:        s.latest,
		LatestByMajor: latestByMajor,
	}
}

func buildVersionsJSON(s parsedStep) spec.Versions {
	entries := make([]spec.VersionEntry, 0, len(s.versionList))
	// Newest-first order: walk versionList in reverse.
	for i := len(s.versionList) - 1; i >= 0; i-- {
		v := s.versionList[i]
		step := s.versions[v]
		var publishedAt *time.Time
		if step.PublishedAt != nil && !step.PublishedAt.IsZero() {
			publishedAt = step.PublishedAt
		}
		commit := ""
		if step.Source != nil {
			commit = step.Source.Commit
		}
		entries = append(entries, spec.VersionEntry{
			Version:       v,
			PublishedAt:   publishedAt,
			HasExecutable: step.Executables != nil && len(*step.Executables) > 0,
			Commit:        commit,
		})
	}
	return spec.Versions{
		StepID:   s.id,
		Latest:   s.latest,
		Versions: entries,
	}
}

// ---------------------------------------------------------------------------
// writer — tracks file count + byte count for Stats
// ---------------------------------------------------------------------------

// fileWriter abstracts the OS calls used by writeJSON, making them injectable
// for testing without affecting the fs.FS-based read path.
type fileWriter interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
}

type realFileWriter struct{}

func (realFileWriter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
func (realFileWriter) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

type writer struct {
	outputDir string
	fw        fileWriter
	fileCount int
	byteCount int64
}

func (w *writer) writeJSON(relPath string, v any) error {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	full := filepath.Join(w.outputDir, relPath)
	if err := w.fw.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	if err := w.fw.WriteFile(full, bytes, 0o644); err != nil {
		return err
	}
	w.fileCount++
	w.byteCount += int64(len(bytes))
	return nil
}

func (w *writer) copyFileFromFS(srcFS fs.FS, srcPath, relDst string) error {
	dst := filepath.Join(w.outputDir, relDst)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := srcFS.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	n, err := io.Copy(out, in)
	if err != nil {
		return err
	}
	w.fileCount++
	w.byteCount += n
	return nil
}
