# envman

## Who and for what?

- connect tools with each other : one tool saves an ENV, the other uses it (input & output)
- manage your environment-sets : use different ENVs for different projects


## How?


## TODO

- `init`: create an empty .envstore file into the current directory
- multi ENV file handling
  - with an arg: `-envstore=path/to/envstore/file.yml` : use this file
  - if there's a .envstore file in the current dir use that one
  - if neither is present: use `$HOME/.envman/.envstore`
- check Go standard package errors (ex: `os.Setenv`)
- expand environment variables **on read** (not at store)
- better progress feedback:
  - present the file path the env is saved into for `add` command
- better command error handling
- store ENVs as Map, not as Slice/array
- better help texts
- **print**: should work for empty as well
- clear : empty the store
