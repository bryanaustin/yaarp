package yaarp

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// FlagSet is the yaarp wrapper for parsing flags.
type FlagSet struct {
	*flag.FlagSet
	parsed bool
	args   []string
}

// BoolFlagValue represents the special boolean value in the flag library
type BoolFlagValue interface {
	flag.Value
	IsBoolFlag() bool
}

var (
	ErrOptionNotFlag  = errors.New("used as a flag when it expects a value")
	ErrOptionNotFound = errors.New("option not found")
)

// CommandLine is the default set of command-line flags, parsed from os.Args.
var CommandLine = &FlagSet{
	FlagSet: flag.CommandLine,
}

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	CommandLine.Parse(os.Args[1:])
}

// Parsed reports whether the command-line flags have been parsed.
func Parsed() bool {
	return CommandLine.Parsed()
}

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// Arg returns the i'th command-line argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func Arg(i int) string {
	return CommandLine.Arg(i)
}

// NArg is the number of arguments remaining after flags have been processed.
func (f *FlagSet) NArg() int { return len(f.args) }

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int { return len(CommandLine.args) }

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string { return f.args }

// Args returns the non-flag command-line arguments.
func Args() []string { return CommandLine.args }

// VisitAll not implemented. Does anyone use it?

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	err := f.parseInternal(arguments)
	if err == nil {
		return nil
	}

	if err == flag.ErrHelp {
		if f.FlagSet.Usage == nil {
			if f.FlagSet.Name() == "" {
				fmt.Fprintf(f.FlagSet.Output(), "Usage:\n")
			} else {
				fmt.Fprintf(f.FlagSet.Output(), "Usage of %s:\n", f.FlagSet.Name())
			}
			f.FlagSet.PrintDefaults()
		} else {
			f.FlagSet.Usage()
		}
	}

	switch f.FlagSet.ErrorHandling() {
	case flag.ContinueOnError:
		return err
	case flag.ExitOnError:
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(f.FlagSet.Output(), "Error parsing args: %s\n", err)
		os.Exit(2)
	case flag.PanicOnError:
		panic(err)
	}

	return nil
}

func (f *FlagSet) parseInternal(arguments []string) error {
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		if arg == "" {
			continue
		}
		if arg == "--" {
			f.args = append(f.args, arguments[i+1:]...)
			return nil
		}
		if arg == "-" || arg[0] != '-' {
			f.args = append(f.args, arg)
			continue
		}
		if strings.HasPrefix(arg, "--") {
			consumed, err := f.parseLongOption(arg[2:], arguments[i+1:])
			if err != nil {
				return err
			}
			i += consumed
			continue
		}
		consumed, err := f.parseShortOptions(arg[1:], arguments[i+1:])
		if err != nil {
			return err
		}
		i += consumed
	}
	return nil
}

func (f *FlagSet) parseLongOption(body string, rest []string) (int, error) {
	name, value, hasEq := splitOnEquals(body)
	fo := f.FlagSet.Lookup(name)
	if fo == nil {
		if name == "help" && !hasEq {
			return 0, flag.ErrHelp
		}
		return 0, fmt.Errorf("option %q: %w", name, ErrOptionNotFound)
	}
	if hasEq {
		fo.Value.Set(value)
		return 0, nil
	}
	if isBoolFlag(fo) {
		fo.Value.Set("true")
		return 0, nil
	}
	return consumeNextValue(fo, rest)
}

func (f *FlagSet) parseShortOptions(body string, rest []string) (int, error) {
	flagBody := body
	attachedValue := ""
	hasEq := false
	if eq := strings.IndexByte(body, '='); eq >= 0 {
		flagBody = body[:eq]
		attachedValue = body[eq+1:]
		hasEq = true
	}

	runes := []rune(flagBody)
	if len(runes) == 0 {
		return 0, fmt.Errorf("option %q: %w", "", ErrOptionNotFound)
	}

	for _, r := range runes[:len(runes)-1] {
		name := string(r)
		fo := f.FlagSet.Lookup(name)
		if fo == nil {
			if name == "h" {
				return 0, flag.ErrHelp
			}
			return 0, fmt.Errorf("option %q: %w", name, ErrOptionNotFound)
		}
		if !isBoolFlag(fo) {
			return 0, fmt.Errorf("option %q: %w", name, ErrOptionNotFlag)
		}
		fo.Value.Set("true")
	}

	last := string(runes[len(runes)-1])
	fo := f.FlagSet.Lookup(last)
	if fo == nil {
		if last == "h" && !hasEq {
			return 0, flag.ErrHelp
		}
		return 0, fmt.Errorf("option %q: %w", last, ErrOptionNotFound)
	}
	if hasEq {
		fo.Value.Set(attachedValue)
		return 0, nil
	}
	if isBoolFlag(fo) {
		fo.Value.Set("true")
		return 0, nil
	}
	return consumeNextValue(fo, rest)
}

// consumeNextValue sets fo to the first non-empty element of rest, returning
// how many of rest's leading elements were consumed (skipped empties + 1).
// Empty input returns len(rest), 0 — matches the original silent-failure
// behavior when a value-expecting flag has no value available.
func consumeNextValue(fo *flag.Flag, rest []string) (int, error) {
	for j, a := range rest {
		if a != "" {
			fo.Value.Set(a)
			return j + 1, nil
		}
	}
	return len(rest), nil
}

func splitOnEquals(s string) (name, value string, hasEq bool) {
	if i := strings.IndexByte(s, '='); i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", false
}

func isBoolFlag(fo *flag.Flag) bool {
	bv, ok := fo.Value.(BoolFlagValue)
	return ok && bv.IsBoolFlag()
}
