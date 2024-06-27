package activator

type ActivatedStep struct {
	StepYMLPath string

	// TODO: this is a mess and only makes sense in the context of a path:: ref
	// This should be cleaned up when all step actions are moved here from the CLI,
	// but I don't want to blow up my PR with that change.
	OrigStepYMLPath string

	WorkDir string
}
