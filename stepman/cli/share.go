package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func share(c *cli.Context) {
	fmt.Println(`
To share your step walk througth this steps:

- Fork the steplib repo
- stepman setup -share -c https://github.com/gkiki90/bitrise-steplib.git
  - You can find your step lib repo at: /Users/godrei/.stepman/step_collections/1438870289/collection
- stepman create -c https://github.com/gkiki90/bitrise-steplib.git new-xcode-archive@1.2.0
  - NOTE: mkdir -p ./steps/new-xcode-archive/1.2.0
  - You can find your step's step.yml at: /Users/godrei/.stepman/step_collections/1438870289/collection/steps/new-xcode-archive/1.2.0/step.yml
  - Open this step.yml, fill out the required infos
  - Once you're happy with it and want to share:
    - you can commit your changes in (it's just a regular git repository):
      - cd /Users/godrei/.stepman/step_collections/1438870289/collection
      - git checkout -b new-xcode-archive
      - git add ./steps/new-xcode-archive/1.2.0/step.yml
      - git commit -m 'new-xcode-archive 1.2.0'
      - git push
    - or call: stepman share -c https://github.com/gkiki90/bitrise-steplib.git new-xcode-archive@1.2.0
      to do it automagically
- Create a pull request

You can find a template step repository at: https://github.com/bitrise-io/bitrise-steplib/step-template/step.yml
`)
}
