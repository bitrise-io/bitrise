# Lesson 2 - The flow of work in your Workflow

[Back to Lesson 1](../lesson1_steps)
[Lesson 3 - The ins and outs with environment variables](../lesson3_input_output_env)

Basically Workflows are groups of steps. There are main Workflows, that contain the Steps which provide the main functionality. There are utility Workflows that we use to prepare everything for the main Workflow, to clean up or to send notification containing the build status. The utility Workflows begin with '_' and these Workflows can't be run using the `bitrise run <workflowname>` command.

What could be a better example to show how Workflows work, than to create an iOS Unit Test Workflow? Let's get down to business!
First of all, what do we need in the Unit Test Workflow?
- Xcode: Test is all that we need to run

And what are the needed setup steps to accomplish these objectives, what should be added to the utility Workflows?
- The project should be on the machine running the Workflow, so there should be a git-clone Step
- There should be a notification Step to make sure you don't have to sit in front of your computer and watch the terminal the whole time

So let's create our first utility Workflow called _setup to make sure that the project is present on the current machine and is up-to-date.
We'll use a simple bash script to achieve this (just for the fun of it ;) ) The _setup Workflow should look something like this:

_setup:
  description: Clone repo
  steps:
  - script:
      title: clone
      run_if: |-
        {{enveq "$XCODE_PROJECT_PATH" ""}}
      inputs:
      - content: |-
          #!/bin/bash
          echo $XCODE_PROJECT_PATH
          if [ ! -d $PROJECT_FOLDER ] ; then
            git clone ${REPO_URL}
          else
            cd $PROJECT_FOLDER
            git pull
          fi

Great! Now let's jump to the main Workflow. It will only contain an Xcode: Test step so let's keep it simple and call it `test`. You can add Workflows to the after_run and before_run of Workflow. This will run the given Workflow just before or after the given Workflow. So here is the main Workflow with the before_run and after_run sections:

test:
  before_run:
    - _setup
  after_run:
    - _cleanup
  steps:
  - xcode-test:
      title: Run Xcode test
      inputs:
      - project_path: ${XCODE_PROJECT_PATH}
      - scheme: ${XCODE_PROJECT_SCHEME}
      - simulator_device: iPhone 6
      - simulator_os_version: latest
      - is_clean_build: "no"

Awesome! Now we are almost done! only one more Workflow to create! _cleanup should contain simply be another bash script that just delete's the directory.

_cleanup:
  description: |-
    This is a utility workflow. It runs a script to delete the folders created in the setup.
  steps:
  - script:
      title: Cleanup folder
      description: |-
        A script step to delete the downloaded Step folder.
      inputs:
      - content: |-
          #!/bin/bash
          rm -rf $PROJECT_TITLE

Wow! We're done! Weeell not quite. If you try to run the Workflow you can see, that it fails. Currently the environment variables aren't added that are needed. Add these environment variables to your .bitrise.secrets.yml:

- REPO_URL: <your-repo-url>
- PROJECT_TITLE: <your-project-title>
- PROJECT_FOLDER: <your-project-folder>
- XCODE_PROJECT_PATH: <your-project-path>
- XCODE_PROJECT_SCHEME: ${PROJECT_TITLE}

Aaaaand yeah! All done! Great job! *Drop mic*

[Back to Lesson 1](../lesson1_steps)
[Lesson 3 - The ins and outs with environment variables](../lesson3_input_output_env)
