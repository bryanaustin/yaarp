# YAARP (Yet Another ARgument Parser)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/bryanaustin/yaarp)

A lightweight argument parser that brings [getopt](https://en.wikipedia.org/wiki/Getopt)-style syntax to Go while reusing the standard library's `flag` package wherever possible. Flags are defined with stdlib `flag` exactly as you would normally; yaarp only replaces the parsing step, so it drops into existing `flag`-based programs with no changes other than the call to `Parse`.

Requires Go 1.17+. No external dependencies.

## Install

```bash
go get github.com/bryanaustin/yaarp
```

## Usage

Define your flags with stdlib `flag`, then call `yaarp.Parse()` instead of `flag.Parse()`:

```go
package main

import (
	"flag"
	"fmt"

	"github.com/bryanaustin/yaarp"
)

func main() {
	count := flag.Int("n", 1, "count")
	verbose := flag.Bool("v", false, "verbose")
	format := flag.String("format", "long", "output format")
	yaarp.Parse()

	fmt.Printf("Count:   %d\n", *count)
	fmt.Printf("Verbose: %v\n", *verbose)
	fmt.Printf("Format:  %s\n", *format)
	fmt.Printf("Args:    %v\n", yaarp.Args())
}
```

Read positional arguments through `yaarp.Args()`, `yaarp.Arg(i)`, and `yaarp.NArg()` — not the corresponding methods on `flag`, since yaarp tracks its own positional list.

## Syntax Details

* A bare single hyphen `-` is treated as a positional argument (commonly meaning "stdin").
* A single hyphen followed by a non-hyphen character (e.g. `-a`) is a short option.
* Any rune may be used as an option name, not just alphanumerics — valid: `-_`, `-\x00`, `-☺`.
* Multiple short options can be chained after one dash as long as only the last one takes a value: `-abc=d` is equivalent to `-a -b -c=d`. Every char before the last must be a boolean flag.
* `--` on its own ends option processing — every following token, even one that looks like a flag, becomes a positional argument.
* Equals signs are optional: `--option=1` and `--option 1` are equivalent. The same holds for the trailing short option: `-o=1` and `-o 1`.
* Options and positional arguments may be interleaved: `-x=2 arg1 -y` is valid.
* `-h` and `--help` are intercepted and trigger usage output unless you've defined them as real flags.

## Example

Given the following flag definitions:

```go
silent := flag.Bool("s", false, "silent")
color  := flag.Bool("color", false, "use color")
output := flag.String("o", "", "output file, use '-' for stdout")
yaarp.Parse()
arguments := yaarp.Args()
```

Invoked as:

```bash
program - -so - bravo --color -- charlie --delta
```

results in:

| Argument/Option | Value     |
| --------------- | --------- |
| silent          | true      |
| color           | true      |
| output          | `-`       |
| arguments[0]    | `-`       |
| arguments[1]    | bravo     |
| arguments[2]    | charlie   |
| arguments[3]    | --delta   |

Note that `-so -` parses as `-s` (bool) plus `-o` taking the next token (`-`) as its value, and everything after `--` becomes a positional regardless of leading dashes.

## Custom FlagSet

A `yaarp.FlagSet` embeds a `*flag.FlagSet`, so you can build an isolated parser instead of using the package-global `CommandLine`:

```go
fs := &yaarp.FlagSet{FlagSet: flag.NewFlagSet("mytool", flag.ContinueOnError)}
verbose := fs.Bool("v", false, "verbose")
if err := fs.Parse(os.Args[1:]); err != nil {
	// handle err
}
_ = verbose
```

## Error Handling

`Parse` honors the underlying `*flag.FlagSet`'s `ErrorHandling`:

* `flag.ContinueOnError` — returns the error (`ErrOptionNotFound`, `ErrOptionNotFlag`, or `flag.ErrHelp`) so the caller can decide what to do.
* `flag.ExitOnError` — the default for `flag.CommandLine`. Exits 0 on `-h`/`--help`, exits 2 (after printing `Error parsing args: ...`) on any other error.
* `flag.PanicOnError` — panics with the error.

A short option used in a chain that requires a value returns `ErrOptionNotFlag` — for example, given a bool `-a` and a string `-v`, the input `-avx` errors at `v` because `v` would need to consume `x` as its value but is not in the trailing position.

## Help Output

If you supply your own `Usage` function on the underlying `*flag.FlagSet`, yaarp calls it. Otherwise yaarp prints a GNU-style usage block:

```
Usage: mytool [OPTIONS] [ARGUMENTS]

Options:
  -v               verbose
  --format <long>  output format (default "long")
  -h, --help       show this help and exit
```

`-h, --help` is appended automatically when neither `-h` nor `--help` has been defined as a real flag. String defaults are quoted; boolean defaults are only shown when the default is `true`.
