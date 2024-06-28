package activator

type ActivatedStep struct {
	StepYMLPath string

	// TODO: this is a mess and only makes sense in the context of a path:: ref
	// This should be cleaned up when all step actions are moved here from the CLI,
	// but I don't want to blow up my PR with that change.
	OrigStepYMLPath string

	// TODO: is this always the same as the `activatedStepDir` function param in all activations? Can we clean this up?
	WorkDir string

	// DidStepLibUpdate indicates that the local steplib cache was updated while resolving the exact step version.
	DidStepLibUpdate bool
}
