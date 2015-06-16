# envman

## Who and for what?

- connect tools with each other : one tool saves an ENV, the other uses it (input & output)
- manage your environment-sets : use different ENVs for different projects


## How? - Use cases

- multi PATH handling: you have `packer` in your `$HOME` dir, in a bin subdir and terraform in another
	- create an envman `.envset` to include these in your `$PATH`


## TODO

- ~~`init`: create an empty .envstore file into the current directory~~
- ~~move CLI commands to separate files, one for each command
	- like in [https://github.com/docker/swarm](https://github.com/docker/swarm)~~
- ~~multi ENV file handling~~
	- with an arg: `-envstore=path/to/envstore/file.yml` : use this file
	- if there's a .envstore file in the current dir use that one
	- if neither is present: use `$HOME/.envman/.envstore`
- ~~check Go standard package errors (ex: `os.Setenv`) [!!!os.Setenv removed!!!]~~
- expand environment variables **on read** (not at store)
- ~~better progress feedback:~~
	- present the file path the env is saved into for `add` command
- ~~better command error handling~~
- ~~store ENVs as Map, not as Slice/array~~
- better help texts
- **print**: should work for empty as well
- ~~clear : empty the store~~
- `env [bash/fish]` : exports ENVs in a bash/fish compatible format
	- for bash it prints `export KEY=value` statements
	- which can be run like this: `$(envman env bash)` to import the ENVs into the current ENV session


## Develop & Test in Docker

* Build: `docker build -t envman .`
* Run: `docker run --rm -it -v `pwd`:/go/src/github.com/bitrise-io/envman --name envman-dev envman /bin/bash`

Or use the included scripts:

* To build&run: `bash _scripts/docker_build_and_run.sh`
* Once it's built you can just run it: `bash _scripts/docker_run.sh`


###################################################################################################################
## USAGE
###################################################################################################################

Environment varaibale manager

VERSION:	0.0.1

USAGE:		envman [OPTIONS] COMMAND [arg...]

VARIABLES:

	ENVSTORE 			The .envstore.yml yaml file, wich feeds envman
	ENVMAN_WORK_DIR		The directory, wich contains ENVSTORE
	ENVMAN_WORK_PATH	The file path of ENVSTORE (ENVMAN_WORK_DIR/ENVSTORE)


OPTIONS:

	--path 			Envman's working path. This is file path, with format {SOME_DIR/envstore.yml}

					* Notes:
						- ENVMAN_WORK_PATH


	--help, -h		Show help


  	--version, -v	Print the version


COMMANDS:

	init, i		Create an empty ENVSTORE into ENVMAN_WORK_DIR (i.e. create ENVMAN_WORK_PATH)

				* Usage: 
					- envman [OPTIONS] init
				* Notes:
					- if ENVSTORE exist does not override it
					- ENVMAN_WORK_PATH is specified by flag (--path [SOME_DIR/envstore.yml]), 
					or it is the CURRENT_DIR/.envstore.yml


  	add, a		Add new or update an exist environment variable

  				* Usage: 
  					- envman [OPTIONS] add --key YOUR_KEY --value YOUR_VALUE
  					- SOME_CMD_THAT_GENERATES_STDOUT | envman [OPTIONS] add --key YOUR_KEY 
  				* Flags: 
  					- key : 	the key of the environment variable
  					- value : 	the value of the environment variable
  				* Notes:
	  				- if ENVSTORE does not exist creates it in ENVMAN_WORK_DIR
	  				- if value exist whit specified key, overrides it with the new value
	  				- piped version saves the SOME_CMD_THAT_GENERATES_STDOUT output with the specified key


	clear, c	Clears the envman provided enviroment variables

				* Usage: 
					- envman [OPTIONS] clear
				* Notes:
					- does nothing, if no envstore in ENVMAN_WORK_PATH


	print, p	Prints out the environment variables in ENVMAN_WORK_PATH

				* Usage: 
					- envman [OPTIONS] print

  	
  	run, r		Runs the specified command with environment variables in ENVSTORE

  				* Usage: 
  					- envman [OPTIONS] run [CMD_TO_RUN]
  				* Notes: 
  					- the cmd will fail, if no ENVSTORE exist in ENVMAN_WORK_PATH
  					- CMD_TO_RUN's Stdin is the standerd input
  					- CMD_TO_RUN's Stdout is the standerd output
  					- CMD_TO_RUN's Stderr is the standerd error

  
  	help		Shows a list of commands or help for one command
