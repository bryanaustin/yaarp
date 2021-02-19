# YAARP (Yet Another ARgument Parser) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/bryanaustin/yaarp)
This argument parser seeks to support more [getopt](https://en.wikipedia.org/wiki/Getopt)-like syntax and to be lightweight. Using as much of the `flag` library as possible.

## Usage
```go
package main

import (
	"flag"
	"fmt"
	"github.com/bryanaustin/yaarp"
)

func main() {
	vflag := flag.Bool("v", false, "verbose")
	format := flag.String("format", "long", "output format")
	yaarp.Parse()

	fmt.Printf("Verbose: %v\n", *vflag)
	fmt.Printf("Format: %v\n", *format)
}
```

## Syntax Details
* A single hyphen `-` is an argument.
* A single hyphen `-` followed by a non hyphen (ie. `-a`) is single char argument.
* Any character works for options, not just alphanumeric. Valid: `-_`, `-\x00`, `-☺`
* Multiple single char options can be chined together after single dash as long as only the last one requires an argument. `-abc=d` is equivalent to `-a -b -c=d`.
* Two dashes followed with nothing else after (ie. `--`) end options. 
* Equals signs optional. `--option=1` is the same as `--option 1`
* Options can be specified after arguments. (ie. `-x=2 arg1 -y`)

## Example
```go
silent := flag.Bool("s", false, "slient")
color  := flag.Bool("color", false, "use color")
output := flag.String("o", "", "output file, use '-' for stdout")
yaarp.Parse()
arguments := yaarp.Args()
```
```bash
program - -so - one --color -- two --three
```
Argument/Option | Value
--------------- | -----
silent          | true
color           | true
output          | -
arguments[0]    | -
arguments[1]    | one
arguments[2]    | two
arguments[3]    | --three