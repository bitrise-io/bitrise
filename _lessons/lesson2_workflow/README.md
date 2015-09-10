# Lesson 2 - The flow of work in your Workflow

Basically Workflows are groups of steps. There are main Workflows, that contains the Steps that provide the main functionality and there are utility Workflows that we use to prepare everything for the main Workflow and after to clean up or even send notification containing the build status. The utility Workflows begin with '_' and these Workflows can't be run using the `bitrise run <workflowname>` command.
