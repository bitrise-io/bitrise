## Changes

* New audit commands:
  - 'stepman audit --step-yml <YML_PATH>' for validating local Step, before share process.
    (Usefull to validate step, when you finished the Step development, before creating the release of the step.)
  - 'stepman audit --step-yml <YML_PATH> --before-pr' for validating Step, before share Pull Request.
    (Usefull to run this command before you try to finish your step share (i.e. after 'stepman share create' and before 'stepman share finish'), your Pull Request will be validated, with this audit command.)
* 'stepman share create' now makes sure, that you provided valid git tag of the step.
