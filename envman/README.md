# envman

## Who and for what?

- connect tools with each other : one tool saves an ENV, the other uses it (input & output)
- manage your environment-sets : use different ENVs for different projects


## How? - Use cases

- multi PATH handling: you have `packer` in your `$HOME` dir, in a bin subdir and terraform in another
	- create an envman `.envset` to include these in your `$PATH`


## Develop & Test in Docker

* Build: `docker build -t envman .`
* Run: `docker run --rm -it -v `pwd`:/go/src/github.com/bitrise-io/envman --name envman-dev envman /bin/bash`

Or use the included scripts:

* To build&run: `bash _scripts/docker_build_and_run.sh`
* Once it's built you can just run it: `bash _scripts/docker_run.sh`


## USAGE IN RUBY 

### Add environment variable through envman

```
system( "envman add --key SOME_KEY --value SOME_VALUE" )
```


### Add environment variable through envman, with piped add

```
IO.popen('envman add --key SOME_KEY', 'r+') {|f| 
	f.write(SOME_VALUE) 
    f.close_write
    f.read 
}
```

### Let envman to read environment value from file

```
require 'tempfile'

file = Tempfile.new('FILE_NAME')
file.write(SOME_VALUE)
file.close

system( "envman add --key SOME_KEY --valuefile #{file.path}" )

file.unlink # removes the file
```