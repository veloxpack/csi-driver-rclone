# sid : generate sortable identifiers

## Overview

[![GoDoc](https://godoc.org/github.com/chilts/sid?status.svg)](https://godoc.org/github.com/chilts/sid)
[![Build Status](https://travis-ci.org/chilts/sid.svg?branch=master)](https://travis-ci.org/chilts/sid)
[![Code Climate](https://codeclimate.com/github/chilts/sid/badges/gpa.svg)](https://codeclimate.com/github/chilts/sid)
[![Go Report Card](https://goreportcard.com/badge/github.com/chilts/sid)](https://goreportcard.com/report/github.com/chilts/sid)

This package is simple and only provides one function. The aim here is not pure speed, it is for an easy use-case
without having to worry about goroutines and locking.

## Install

```
go get github.com/chilts/sid
```

## Example

```
id1 := sid.Id()
id2 := sid.IdHex()
id3 := sid.IdBase32()
id4 := sid.IdBase64()

fmt.Printf("id1 = %s\n", id1)
fmt.Printf("id2 = %s\n", id2)
fmt.Printf("id3 = %s\n", id3)
fmt.Printf("id4 = %s\n", id4)

// -> "id1 = 1559872035903071353-1186579057231285506"
// -> "id2 = 15a5cf57e7d2a837-6eaafe687e7b3ec3"
// -> "id3 = 1b9efqnl51jj7-4u66ikpfq9ugm"
// -> "id4 = 1IeSBAWW9kK-0cDG64GQgGJ"
```

## Author

By [Andrew Chilton](https://chilts.org/), [@twitter](https://twitter.com/andychilton).

For [AppsAttic](https://appsattic.com/), [@AppsAttic](https://twitter.com/AppsAttic).

## License

[MIT](https://chilts.mit-license.org/2017/)

(Ends)
