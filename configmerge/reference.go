package configmerge

import "fmt"

type configReference struct {
	Repository string `yaml:"repository" json:"repository"`
	Branch     string `yaml:"branch" json:"branch"`
	Commit     string `yaml:"commit" json:"commit"`
	Tag        string `yaml:"tag" json:"tag"`
	Path       string `yaml:"path" json:"path"`
}

func (r configReference) Key() string {
	var key string
	if r.Repository != "" {
		key = fmt.Sprintf("repo:%s,%s", r.Repository, r.Path)
	} else {
		key = r.Path
	}

	if r.Commit != "" {
		key += fmt.Sprintf("@%s", r.Commit)
	} else if r.Tag != "" {
		key += fmt.Sprintf("@%s", r.Tag)
	} else if r.Branch != "" {
		key += fmt.Sprintf("@%s", r.Branch)
	}

	return key
}
