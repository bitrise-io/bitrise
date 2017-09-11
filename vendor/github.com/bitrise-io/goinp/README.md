# go-interactive-cli

Interactive CLI in Go (test)

## Ask for an input from the user with the `AskForXyz` methods.

Ask for a string input with `AskForString`

Ask for a 64 bit integer (int64) input with `AskForInt`

Ask for a bool input with `AskForBool`

* this method accepts all the standard true/false values handled by [http://golang.org/pkg/strconv/#ParseBool](http://golang.org/pkg/strconv/#ParseBool)
    * this includes: `1`, `t`, true`, `0`, `f`, false`
* additionally `yes`, `y`, `no` and `n` are also accepted
* every input handled in a case insensitive way, so `TrUe` will also return `true`
