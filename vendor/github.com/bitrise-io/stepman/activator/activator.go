package activator

type ActivatedStep struct {
	StepYMLPath string

	// ExecutablePath is a local path to the main entrypoint of the step, ready for execution.
	// This can be an empty string if:
	// - step was activated from a git reference (we checked out the source dir directly)
	// - step was activated from a local path (we copied the source dir directly)
	// - step was activated from a steplib reference, but step.yml has no entry for pre-compiled binaries (we fallback to source checkout)
	// - step was activated from a stpelib reference, but step.yml has no pre-compiled binary for the current OS+arch combo (we fallback to source checkout)
	ExecutablePath string

	// DidStepLibUpdate indicates that the local steplib cache was updated while resolving the exact step version.
	// TODO: this is a leaky abstraction and we shouldn't signal this here, but it requires a bigger refactor.
	// (stepman should keep track of this info in a file probably)
	DidStepLibUpdate bool
}
