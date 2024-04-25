# go-onigmo ![Test](https://github.com/go-enry/go-oniguruma/workflows/Test/badge.svg)

go-onigmo is a drop-in replacement of the `regexp` package from the Go standard library. Onigmo is the default  regexp library in Ruby 2.0.

As differentiation factor from many other `oniguruma` or `onigmo` wrappers we follow the next goals:

- Be thread-safe; this library has been tested in high concurrency environments.
- Implement the same interface that `regexp.Regexp` does; all the method from it has been implemented. With small exceptions.
- Implement the same behavior of regexp.Regexp`; this library is tested against the test suite from the standard library. All tests pass with some exceptions.
- Provide easy mechanic to enable/disable this functionally, and avoid to force all the dependant library to install always the dependencies.


Limitations
-----------

The goals of this package are to archive full compatibility with the standard library. Still, due to the limitation of onigmo, and the uniqueness of the Go implementation some times is hard.

These are the mismatches between this library and the standard library `regexp` package:
:

- `Regexp.LiteralPrefix` it's not implemented.
- Expressions with duplicate named aren't supported. `(?P<x>hi)|(?P<x>bye)`
- Nested repetition operators are supported, such as `a**` or `a*+`.

Install
-------

```sh
# linux (debian/ubuntu/...)
sudo apt-get install libonig-dev

# osx (homebrew)
brew install oniguruma

go get github.com/go-enry/go-oniguruma
```

Attributions
------------

This project it's based on the [work](https://github.com/moovweb/rubex/tree/go1) of Zhigang Chen <zhigangc@gmail.com>.  

License
-------
Apache License Version 2.0, see [LICENSE](LICENSE)
