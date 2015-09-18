# Lesson 4 - Keeping the control even when the engines are on fire a.k.a. Error management

[Back to Lesson 3](../lesson3_input_output_env)

[Lesson 5 - A complex Workflow](../lesson5_complex_wf)

So here's one of the most common part of development - Errors. Yeah we all know that guy who configures everything and writes every line of code flawlessly at the first time... Of course we are not that guy. When working on a complex workflow, it happens at least once that an old code stays in the project making the tests fail or that one little option in the configuration that messes up the whole thing and makes it fail.

We are following the bash conventions about error handling. Every Step that runs successfully exits with the error code 0 and if the exit code is different the Step fails and (if the Step wasn't skippable the whole Workflow fails).

There are two ways to keep the Workflow up and running even after a failed Step.
- If the Step was marked skippable, the following Steps will also run. This is great if you want to notify the team that the build started but the used service is currently offline.
- If the Step is marked always run the given Step will be run even if the build fails. This can be used to notify the team of errors.

[Back to Lesson 3](../lesson3_input_output_env)

[Lesson 5 - A complex Workflow](../lesson5_complex_wf)
