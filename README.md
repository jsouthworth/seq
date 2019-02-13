# Seq

[![GoDoc](https://godoc.org/jsouthworth.net/go/seq?status.svg)](https://godoc.org/jsouthworth.net/go/seq)
[![Build Status](https://travis-ci.org/jsouthworth/seq.svg?branch=master)](https://travis-ci.org/jsouthworth/seq)
[![Coverage Status](https://coveralls.io/repos/github/jsouthworth/seq/badge.svg?branch=master)](https://coveralls.io/github/jsouthworth/seq?branch=master)

Seq is a lazy sequence library for go. It is inspired by Clojure's sequence functions. It originally started because I was interested in transducers and wanted to play with them. Most of the functions in the library are implemeneted as transducers in the [jsouthworth.net/go/transduce](https://godoc.org/jsouthworth.net/go/transduce) library and wrapped with a XfrmSequence here. This library relies heavily on reflection to allow for the most flexibility in what the user provides. This means that it is only type checked at runtime.

## Getting started
```
go get jsouthworth.net/go/seq
```

## Usage

The full documentation is available at
[jsouthworth.net/go/seq](https://jsouthworth.net/go/seq)

## License

This project is licensed under the MIT License - see [LICENSE](LICENSE)

## Acknowledgments

* Clojure's sequence library was the source of inspiration for the organization of this library and the names of the functions.

## TODO

* [ ] Implement a few more functions like flatten.
* [ ] Add more conversions for go types (maps as entry sequences, etc.).

