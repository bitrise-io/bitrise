# envman

## Who and for what?

- connect tools with each other : one tool saves an ENV, the other uses it (input & output)
- manage your environment-sets : use different ENVs for different projects
- complex environment values : if you want to store a complex input as ENV (for example a change log text/summary), `envman` makes this easy, you don't have to encode the value so that when you call bash's `export KEY=value` it will store your value as you intended. You can just give `envman` the value as it is through a `-valuefile` option or as an input stream (or just edit the related `.envstore` file manually), no encoding required.
- switch between environment sets : if you work on multiple projects where each one requires different environments you can manage this with `envman`


## How? - Use cases

- multi PATH handling: you have `packer` in your `$HOME` dir, in a bin subdir and terraform in another
	- create an envman `.envset` to include these in your `$PATH`


## Develop & Test in Docker

* Build: `docker build -t envman .`
* Run: `docker run --rm -it -v `pwd`:/go/src/github.com/bitrise-io/envman --name envman-dev envman /bin/bash`

Or use the included scripts:

* To build&run: `bash _scripts/docker_build_and_run.sh`
* Once it's built you can just run it: `bash _scripts/docker_run.sh`


---
## Usage example: Ruby

### Add environment variable with `--value` flag

```
system( "envman add --key SOME_KEY --value SOME_VALUE" )
```

### Add environment variable from an input stream

```
IO.popen('envman add --key SOME_KEY', 'r+') {|f| 
	f.write(SOME_VALUE) 
	f.close_write
	f.read 
}
```

### Add environment variable with a value file

```
require 'tempfile'

file = Tempfile.new('SOME_FILE_NAME')
file.write(SOME_VALUE)
file.rewind

system( "envman add --key SOME_KEY --valuefile #{file.path}" )

file.close
file.unlink 
```
  
---
## Usage exmale: GO 

### Add environment variable with `--value` flag

```
import "os/exec"

cmd := "envman add --key SOME_KEY --value SOME_VALUE"
c := exec.Command("bash", "-c", cmd)
err := c.Run()
if err != nil {
   // Handle error
}
```

### Add environment variable from an input stream

```
import "os/exec"

cmd := "echo SOME_VALUE | envman add --key SOME_KEY"
c := exec.Command("bash", "-c", cmd)
err := c.Run()
if err != nil {
	// Handle error
}
```

### Add environment variable with a value file

```
import (
	"os/exec"
	"fmt"
)

cmd := fmt.Sprintf("envman add --key SOME_KEY --valuefile %v", SOME_FILE_PATH)
c := exec.Command("bash", "-c", cmd)
err := c.Run()
if err != nil {
	// Handle error
}
```

---
## Usage exmaple: Bash

### Add environment variable with `--value` flag

```
envman add --key SOME_KEY --value SOME_VALUE
```

### Add environment variable from an input stream

```
echo SOME_VALUE | envman add --key SOME_KEY
```

### Add environment variable with a value file

```
envman add --key SOME_KEY --valuefile SOME_FILE_PATH
```
  
---
