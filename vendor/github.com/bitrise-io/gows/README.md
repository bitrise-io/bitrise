# gows

Go Workspace / Environment Manager, to easily manage the Go Workspace during development.


## The idea

Work in **isolated (development) environment** when you're working on your Go projects.
**No cross-project dependency version missmatch**, no more packages left out from `vendor/`.

No need for initializing a go workspace either, **your project can be located anywhere**,
not just in a predefined `$GOPATH` workspace.
`gows` will take care about crearing the (per-project isolated) workspace directory
structure, no matter where your project is located.

`gows` **works perfectly with other Go tools**, all it does is it ensures
that every project gets it's own, isolated Go workspace and sets `$GOPATH`
accordingly.


## Install

### Requirements

* `Go` with a prepared `$GOPATH`
  * required for creating the `GOPATH/bin` symlink, to have shared `GOPATH/bin` between workspaces


### Install & Prepare Go

Install & configure `Go` - [official guide](https://golang.org/doc/install).

Make sure you have `$GOPATH/bin` in your `$PATH`, e.g. by adding

```
export PATH="$PATH:$GOPATH/bin"
```

to your `~/.bash_profile` or `~/.bashrc` (don't forget to `source` it or to open a new Terminal window/tab
if you just added this line to the profile file).

*This makes sure that the Go projects you build/install will be available in any Terminal / Command Line
window, without the need to type in the full path of the binary.*


### Install `gows`

```
go get -u github.com/bitrise-tools/gows
```

That's all. If you have a "properly" configured Go environment (see the previous Install section)
then you should be able to run `gows -version` now, and be able to run `gows` in any directory.


## Usage

Run `gows init` inside your Go project's directory (*you only have to do this once*).
This will initialize the isolated
Workspace for the project and create a `gows.yml` file in your
project's directory (you can commit or `.gitignore` this file if you want to).

Once you initialized the workspace,
just prefix your commands (any command) with `gows`.

Example:

Instead of `go get ./...` use `gows go get ./...`,
instead of `go install` use `gows go install`, etc.

That's pretty much all :)

If you'd want to clean up the workspace you can run `gows clear`
in your project's directory, that'll delete and re-initialize the
related workspace.


### Alternative usage option: jump into a prepared Shell

*This solution works for most shells, but there are exceptions, like `fish`.
The reason is: `gows` creates a symlink between your project and the isolated workspace.
In Bash and most shells if you `cd` into a symlink directory (e.g. `cd my/symlink-dir`)
your `pwd` will point to the symlink path (`.../my/symlink-dir`),
but a few shells (`fish` for example) change the path to the symlink target path instead,
which means that when `go` gets the current path that will point to your project's original
path, instead of to the symlink inside the isolated workspace. Which, at that point,
is outside of GOPATH.*

In shells which keep the working directory path to point to the symlink's path, instead
of it's target (e.g. Bash) you can run:

```
gows bash
```

or

```
gows bash -l
```

which will start a shell (in this example Bash) with prepared `GOPATH` and your
working directory will be set to the symlink inside the isolated workspace.

If you want to use this mode you'll have to change how you initialize
your `GOPATH`, to allow it to be overwritten by `gows` for "shell jump in".
To allow `gows` to overwrite the `GOPATH` for shells initialized **by/through** `gows` you should
change your `GOPATH` init entry in your `~/.bash_profile` / `~/.bashrc` (or
wherever you did set GOPATH for your shell).
For Bash (`~/.bash_profile` / `~/.bashrc`) you can use this form:

```
if [ -z "$GOPATH" ] ; then
  export GOPATH="/my/go/path"
fi
```

instead of this one:

```
export GOPATH="/my/go/path"
```

This means that your (Bash) shell will only set the `GOPATH` environment if it's not set to a value already.

This is not required if you use `gows` only in a "single command / prefix" style,
it's only required if you want to initialize
the shell and jump into the initialized shell through `gows`. In general it's safe to initialize the
environment variable this way even if you don't plan to initialize any shell through `gows`,
as this will always initialize `GOPATH` *unless* it's already initialized (e.g. by an outer shell).


### `gows` commands

*You can get the list of available commands by running: `gows --help`,
and command specific help by running: `gows COMMAND --help`*

* `gows version` : Print the version of `gows`, same as `gows --version`.
* `gows init [--reset] [go-package-name]` : Initialize a workspace for the current directory.
  * If called without a go-package-name parameter `gows` will try to determine the package name
    from `git remote` (`git remote get-url origin`).
  * For more help see: `gows init --help`.
* `gows workspaces` : List registered gows projects -> workspaces path pairs


## Technical Notes, how `gows` works behind the scenes

When you call `gows init` in your project's directory (wherever it is),
`gows` creates an empty Go Workspace for it in `~/.bitrise-gows/wsdirs/`,
and registers your project's path in `~/.bitrise-gows/workspaces.yml`, so
that the same workspace (inside `~/.bitrise-gows/wsdirs/`) can be assigned
for it every time.

When you run any `gows` command from your project's directory, `gows` will
symlink the project directory into the related `~/.bitrise-gows/wsdirs/...`
Workspace directory before running the command. Additionally `gows`
will symlink your original `GOPATH/bin` into the workspace in
`~/.bitrise-gows/wsdirs/...`, so that if you `go install` something that'll
create the binary in your `GOPATH/bin`, not just inside the isolated Workspace.

Once the symlinks are in place `gows` will also set two environments for the command,
`GOPATH` and `PWD`, to point to the isolated workspace and the project path
inside it.

A step by step example about how the directory structure is built:

```
$ cd $GOPATH/src/github.com/bitrise-tools/gows

$ ls -alh ~/.bitrise-gows/
ls: ~/.bitrise-gows/: No such file or directory

$ gows init
...
Successful init - gows is ready for use!

$ tree -L 5 ~/.bitrise-gows/wsdirs/
~/.bitrise-gows/wsdirs/
└── gows-1464900642
    └── src

2 directories, 0 files

$ ls -l1 ~/.bitrise-gows/
workspaces.yml
wsdirs

# the first `gows` command you run creates the symlinks
# inside the related workspace, in `~/.bitrise-gows/wsdirs/`
$ gows pwd
~/.bitrise-gows/wsdirs/gows-1464900642/src/github.com/bitrise-tools/gows

$ tree -L 5 ~/.bitrise-gows/wsdirs/
~/.bitrise-gows/wsdirs/
└── gows-1464900642
    ├── bin -> ~/develop/go/bin
    └── src
        └── github.com
            └── bitrise-tools
                └── gows -> ~/develop/go/src/github.com/bitrise-tools/gows
```


## TODO

- [x] Setup the base code (generate the template project, e.g. create a new Xcode project or `rails new`)
  - [x] commit & push
- [ ] Add linter tools
  - go:
    - [ ] `go test`
    - [ ] `go vet`
    - [ ] [errcheck](github.com/kisielk/errcheck)
    - [ ] [golint](github.com/golang/lint/golint)
- [ ] Write tests & base functionality, BDD/TDD preferred
- [ ] Setup continuous integration (testing) on [bitrise.io](https://www.bitrise.io)
- [ ] Setup continuous deployment for the project - just add it to the existing [bitrise.io](https://www.bitrise.io) config
- [ ] Use [releaseman](https://github.com/bitrise-tools/releaseman) to automate the release and CHANGELOG generation
- [ ] Iterate on the project (and on the automation), test the automatic deployment

- [ ] Add (r)sync mode as an option - as a workaround for "shell jump in" usage mode in Shells which don't work with the symlink based mode (e.g. `fish`)
  - allow it to be specified in the `gows.yml`
  - add commands: `sync-in` and `sync-back`, in case you want to sync with an already open shell (e.g. you changed the code in the Project dir),
    and to be able to `sync-back` in case you missed to add the `-sync-back` flag to the original command
- [ ] Option to disable `GOPATH/bin` symlinking. Once this
  - [ ] configurable in `gows.yml`
  - [ ] it should be able to handle if the user changes the option - should remove the symlink / dir and create the dir / symlink instead