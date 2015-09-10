# Lesson 1 - The first Steps

First of all let's talk about Steps. Steps are the building blocks of a [Bitrise](https://bitrise.io) Workflow. At [Bitrise](https://bitrise.io) we know how important it is to have plenty of opportunities to customize the automation process as we also worked as a mobile application development company. The need for a wide variate of customization in the automation of the development and deployment workflows is what lead to the creation of [Bitrise](https://bitrise.io). We are eager to provide you with Steps that can help you with automation throughout the application lifecycle and this is where our Step Library comes into the picture. We created the open source [StepLib](https://github.com/bitrise-io/bitrise-steplib) to give you the basic Steps that you could need in creating a Workflow to boost your productivity. Yes, it is open source, you can fork it, add your own Steps or even use another public fork of it! Also when you add a useful Step to your fork and you think other developers could make good use of it feel free to send us a pull request!

Now that you created your first local project (by calling the `bitrise setup` and after it the `bitrise init`) we can have some fun with the Steps! Open the bitrise.yml and let's add some steps!

## StepID
SetpID is a unique identifier of a step. In your Workflow you have to include this ID to tell [Bitrise](https://bitrise.io) which Step you'd like to run. In our [StepLib](https://github.com/bitrise-io/bitrise-steplib) if you open the [steps folder](https://github.com/bitrise-io/bitrise-steplib/tree/master/steps) you can see that every Step folder's name is the StepID.

### StepID format in the yml

- For Steps from the [StepLib](https://github.com/bitrise-io/bitrise-steplib)
  - `https://github.com/bitrise-io/bitrise-steplib.git::script@0.9.1:`
    - This is the full StepID format: <step-lib-source>::StepID@version:
  - `::script@0.9.0:` and `script@0.9.0:`
    - If the `default_step_lib_source` is defined (by default it is and refers to our [StepLib](https://github.com/bitrise-io/bitrise-steplib)), you can simply omit the <step-lib-source> and even the `::` separator.
  - `script@:` and `script:`
    - If there is only one version of a step or if you always want to use the latest version you can even remove the version and the `@` separator too. And if you take a look at the generated bitrise.yml you can see that this is the format it uses (the only step in the Workflow is `- script:`)
- For Steps that are not in the [StepLib](https://github.com/bitrise-io/bitrise-steplib) and are stored online
  - The format to download and run a step is git::<clone-url>@<branch>
    - `git::https://github.com/bitrise-io/steps-timestamp.git@master`
      - In this case we are using the HTTPS clone url to clone the master branch in the Step's repository
    - `git::git@github.com:bitrise-io/steps-timestamp.git@master`
      - In this case we are using the SSH clone url to clone the master branch in the Step's repository
- For Steps on your machine
  - In this case the Step is already stored on your computer and we only need to know the exact path to the step.sh
    - relative-path::./steps-timestamp
    - path::~/develop/go/src/github.com/bitrise-io/steps-timestamp
    - path::$HOME/develop/go/src/github.com/bitrise-io/steps-timestamp
