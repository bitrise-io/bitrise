# envman

## Who and for what?

- connect tools with each other : one tool saves an ENV, the other uses it (input & output)
- manage your environment-sets : use different ENVs for different projects


## How?


## TODO

- multi ENV file handling
- check Go standard package errors (ex: `os.Setenv`)
- expand environment variables **on read** (not at store)
- better progress feedback:
  - present the file path the env is saved into for `add` command
- better command error handling
- ~~store ENVs as Map, not as Slice/array~~
- better help texts 
- **print**: should work for empty as well
- clear : empty the store
