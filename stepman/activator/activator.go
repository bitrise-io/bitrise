package activator

type ActivatedStep struct {
	StepYMLPath string

	// DidStepLibUpdate indicates that the local steplib cache was updated while resolving the exact step version.
	DidStepLibUpdate bool
}
