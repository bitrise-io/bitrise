# Lesson 6 - Pull the trigger on the Workflow

Using Git Flow you have multiple branches and need to do different things according to these branch types. Let's try the triggers with an example:
There are some developers working on your project. Each one of them works on a different feature branch developing different features. When a developer finishes a feature and merges the given branch, you want to notify the lead developer that it's time for a code review. When a feature set is merged on the development branch you may want to add the changes to the master branch, deploy the application and send notification emails to some employees of the client. Triggers can be added to the trigger map section in your bitrise.yml. You set a pattern and which workflow should the given pattern trigger. Here is a sample trigger map for the example development process:

  trigger_map:
  - pattern: test**
    is_pull_request_allowed: true
    workflow: test
  - pattern: "**feature**"
    is_pull_request_allowed: true
    workflow: feature
  - pattern: "**develop"
    is_pull_request_allowed: true
    workflow: develop
  - pattern: master
    is_pull_request_allowed: true
    workflow: master
  - pattern: "*"
    is_pull_request_allowed: true
    workflow: fallback

You can notice that there is a fallback workflow at the end of the trigger map. This Workflow runs if the trigger expression didn't match any of the defined trigger patterns. For example if a developer creates a new branch with the name `develop_awesome_important_change` it wouldn't match the `**develop` trigger pattern. In this case the fallback Workflow would run. You can use this Workflow to get notified about the wrong branch name. As you can see you can add wildcard to your pattern but make sure to add the `""` if you want to start the pattern with the wildcard (in yml the value can't start with *).

You can notice on the [Bitrise website](https://bitrise.io) that the triggers there are the names of the branch that received the push or pull request.

You can try the samples in the bitrise.yml. Just run the `bitrise trigger` command to view the full list of triggers in the .yml and try running the given workflow with the `bitrise trigger <selected_trigger_expression>` command.
