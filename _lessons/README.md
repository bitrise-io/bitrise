# Welcome to Bitrise CLI

Captain's log, bitdate 110912.5.
We begin our mission discovering a new universe, full of new opportunities and new ways to improve our selves and our day-to-day routines. We prepared for this day for a long time.

- The first step was a simple command to make sure we have what it takes to start our adventures: `curl -L https://github.com/bitrise-io/bitrise/releases/download/1.0.0/bitrise-$(uname -s)-$(uname -m) > /usr/local/bin/bitrise`

- Next we made sure that we are in the right mode `chmod +x /usr/local/bin/bitrise`

- And finally we made sure that everything is present and we are using the latest technologies by running the `chmod +x /usr/local/bin/bitrise` command and after it

A journey never had such an easy start. We just traveled to the planet (or if you prefer folder) typed `bitrise init` in our computer and a new Workflow was created that we could use right away to automate a part of our day-to-day routine. We learned a lots of things on our voyage and we are here to help you get started in this automated universe. The lessons section is all about getting familiar with the how-tos of the [Bitrise CLI](https://github.com/bitrise-io/bitrise). Every lesson folder contains a README.md that gives you an overview of the topic and a bitrise.yml that has a complete Workflow ready to run.

- Explore the Steps (including the Steps in our [StepLib](https://github.com/bitrise-io/bitrise-steplib)) in [lesson1](./lesson1_steps)
- Create an army of Steps by adding them to your Workflow to conquer your automation needs in [lesson2](./lesson2_workflow)
- Make sure your army of Steps get and pass on to each other the right orders in [lesson3](./lesson3_input_output_env)
- Stay in control even in hard times (due to errors) in [lesson4](./lesson4_errors)
- And take a look at one of our journey through a complete Workflow in [lesson5](./lesson5_complex_wf)
