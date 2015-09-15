# Lesson 3 - The ins and outs with environment variables

<div style="text-align: left;display: inline-block; width: 50%;">[Back to Lesson 2](../lesson2_workflow)</div><div style="text-align: right;display: inline-block; width: 50%;">[Lesson 4 - Keeping the control even when the engines are on fire](../lesson4_errors)</div>

You are probably familiar with environment variables. These are crucial part of [Bitrise](https://bitrise.io), because our Steps communicate using Environment Variables. We created [envman](https://github.com/bitrise-io/envman) to make Environment Variable management a whole lot easier. Also for security reasons we added a .bitrise.secrets.yml to store all your secret passwords and any other local machine- or user related data. At every `bitrise init` we create a .gitignore file to make sure that the top secret data you are storing in this file is not added to git.

There are multiple ways to create Environment Variables

- You can add them to the `.bitrise.secrets.yml` - these variables will be accessible throughout the whole app (every Workflow).
- You can add them to the envs section of the app, just like the BITRISE_PROJECT_TITLE and BITRES_DEV_BRANCH - these variables will be accessible throughout the whole app (every Workflow).
- You can add them to the envs section of the given Workflow you would like to use it in - these variables will be accessible throughout the Workflow.
- You can export them in your own Workflow by using the [script step from the StepLib](https://github.com/bitrise-io/bitrise-steplib/tree/master/steps/script) -
  - or to make it visible in the whole Workflow you can use [envman](https://github.com/bitrise-io/envman) (`envman add --key SOME_KEY --value 'some value'`)

<div style="text-align: left;display: inline-block; width: 50%;">[Back to Lesson 2](../lesson2_workflow)</div><div style="text-align: right;display: inline-block; width: 50%;">[Lesson 4 - Keeping the control even when the engines are on fire](../lesson4_errors)</div>
