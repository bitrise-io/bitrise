package models

type ActivatedStep struct {
	StepYMLPath string
	ExecutablePath string
	Args []string
	// TODO: envs?
	// TODO: workdir?
}
