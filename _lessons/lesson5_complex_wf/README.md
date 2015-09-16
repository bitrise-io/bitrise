# Lesson 5 - A complex Workflow

Let's spice things up a little bit with a more complex bitrise.yml. We will create a Workflow for an iOS project just like [lesson2](../lesson2_workflow) but this time we'll prepare it for running on our local machine and also on [Bitrise](https://bitrise.io) (Yeah, just for fun we'll run different Steps locally and on the CI server) also we'll add some more main Workflows so that we can use the Archive, Analyze and Test features of Xcode and combine these into a single Workflow by using the before_run after_run fields.

First of all let's summarize what we want.
- Utility
  - _setup
    - git clone or pull to get the source code on the local machine
  - _cleanup
    - remove source from the local machine
  - _download_certs
    - to download the needed certificates on the CI Server
- Main Workflows
  - analyze
  - archive
  - test
  - master - to create the archive, deploy it and notify the Users about it

Move the Workflow from [lesson2](../lesson2_workflow) to the current bitrise.yml. Now we have a _setup, _cleanup and a test Workflow.

Let's add the _download_certs Workflow. It will only have one step, the certificate-and-profile-installer. We have to pass it two inputs - keychain_path and keychain_password. These are the only two parameters that we'll need. We also want to set it to run only on the CI server so we have to set the run_if to .IsCI.
The Workflow should look something like this:

  _download_certs:
    description: This is a utility workflow, used by other workflows.
    summary: This workflow downloads the needed certificates on the CI server and adds them to the keychain.
    steps:
    - git::https://github.com/bitrise-io/steps-certificate-and-profile-installer.git@master:
        description: |-
          This step will only be used in CI mode, on continuous integration
          servers / services (because of the `run_if` statement),
          but **NOT** when you run it on your own machine.
        run_if: .IsCI
        inputs:
        - keychain_path: $BITRISE_KEYCHAIN_PATH
        - keychain_password: $BITRISE_KEYCHAIN_PASSWORD

Now we should add the remaining two Xcode Workflows. For both Workflows the _setup and _download_certs Workflows have to be added to the before_run section, to make sure the source is on the machine and the needed signing tools are also present. The only difference between these two Workflows is that before the archive is created we want to run a Unit Tests to make sure nothing went wrong since the previous deployed version.

  analyze:
    before_run:
    - _setup
    - _download_certs
    description: |-
      This workflow will run Xcode analyze on this project,
      but first it'll run the workflows listed in
      the `before_run` section.
    steps:
    - script:
        title: Run Xcode analyze
        inputs:
        - content: xcodebuild -project "${XCODE_PROJECT_PATH}" -scheme "${XCODE_PROJECT_SCHEME}"
            analyze
  archive:
    description: |-
      This workflow will run Xcode archive on this project,
      but first it'll run the workflows listed in
      the `before_run` section.
    before_run:
    - _setup
    - test
    - _download_certs
    steps:
    - xcode-archive:
        title: Run Xcode archive
        inputs:
        - project_path: ${XCODE_PROJECT_PATH}
        - scheme: ${XCODE_PROJECT_SCHEME}
        - output_dir: $output_dir
        outputs:
        - BITRISE_IPA_PATH: null
          opts:
            title: The created .ipa file's path

And now the master Workflow. This Workflow will deploy the created archive, clean up and send a notification to slack. So the before_run section should contain the archive and the after_run should contain the _cleanup Workflow. And just to make it sure no one uploads a broken version we will set the run_if to only run the Steps if the build is running on a CI server. By adding the correct input variables the Workflow should look like this:

master:
  description: |-
    This workflow is meant to be used on a CI server (like bitrise.io), for continuous
    deployment, but of course you can run it on your own Mac as well,
    except the Step which deploys to Bitrise.io - that's marked with
    a Run-If statement to be skipped, unless you run bitrise in --ci mode.
  before_run:
  - archive
  after_run:
  - _cleanup
  steps:
  - script:
      inputs:
      - content: |-
          #!/bin/bash
          echo "-> BITRISE_IPA_PATH: ${BITRISE_IPA_PATH}"
  - bitrise-ios-deploy:
      description: |-
        The long `run_if` here is a workaround. At the moment Bitrise.io
        defines the BITRISE_PULL_REQUEST environment
        in case the build was started by a Pull Request, and not the
        required PULL_REQUEST_ID - so we'll check for that instead.
      run_if: enveq "BITRISE_PULL_REQUEST" "" | and .IsCI
      inputs:
      - notify_user_groups: none
      - is_enable_public_page: "yes"
      outputs:
      - BITRISE_PUBLIC_INSTALL_PAGE_URL: null
        opts:
          title: Public Install Page URL
          description: |-
            Public Install Page's URL, if the
            *Enable public page for the App?* option was *enabled*.
  - slack:
      run_if: .IsCI
      inputs:
      - webhook_url: ${SLACK_WEBHOOK_URL}
      - channel: ${SLACK_CHANNEL}
      - from_username: ${PROJECT_TITLE} - OK
      - from_username_on_error: ${PROJECT_TITLE} - ERROR
      - message: |-
          CI check - OK
          PULL_REQUEST_ID : ${PULL_REQUEST_ID}
          BITRISE_PUBLIC_INSTALL_PAGE_URL: ${BITRISE_PUBLIC_INSTALL_PAGE_URL}
      - message_on_error: |-
          CI check - FAILED
          PULL_REQUEST_ID : ${PULL_REQUEST_ID}
