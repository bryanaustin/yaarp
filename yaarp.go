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

const (
	stateDefault = iota
	stateBufArgument
	stateOptionStart
	stateDoubleDash
	stateLongOption
	stateShortOptions
	stateValueExpected
	stateArgumentOnly
)

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
		os.Exit(2)
	case flag.PanicOnError:
		panic(err)
	}

	return nil
}

func (f *FlagSet) parseInternal(arguments []string) error {
	var state, i1, i2 int
	var option string
	buffer := &strings.Builder{}

	trySetBool := func(fo *flag.Flag) (valueset bool) {
		if bv, ok := fo.Value.(BoolFlagValue); ok && bv.IsBoolFlag() {
			bv.Set("true")
			valueset = true
		}
		return
	}

	for ; ; i2++ {
		var seperator bool
		var focus rune
		var runeArgs []rune

		runeArgs = []rune(arguments[i1])

		if i2 == len(runeArgs) {
			seperator = true
		} else {
			if i2 > len(runeArgs) {
				i1++
				if i1 >= len(arguments) {
					break
				}
				i2 = 0
			}
			runeArgs = []rune(arguments[i1])
			focus = runeArgs[i2]
		}

		// if seperator {
		// 	fmt.Println(fmt.Sprintf("%d separator", state))
		// } else {
		// 	fmt.Println(fmt.Sprintf("%d %q", state, focus))
		// }

		switch state {

		// Anything could happen next!
		case stateDefault:
			if focus == '-' {
				state = stateOptionStart
			} else if !seperator {
				buffer.WriteRune(focus)
				state = stateBufArgument
			}

		// Receiving an argument (not an option), keep buffering
		// until we come across a separator
		case stateBufArgument:
			if seperator {
				f.args = append(f.args, buffer.String())
				buffer.Reset()
				state = stateDefault // argument captured, return to default state
			} else {
				buffer.WriteRune(focus)
			}

		// This will probably be an option. There are other situations were it
		// not be, but it probably is.
		case stateOptionStart:
			if seperator {
				f.args = append(f.args, "-")
				state = stateDefault // add a -, return to default state
			} else if focus == '-' {
				state = stateDoubleDash
			} else {
				buffer.WriteRune(focus)
				state = stateShortOptions
			}

		// Two dashes happened, is it going to be a long option?
		case stateDoubleDash:
			if seperator {
				state = stateArgumentOnly
			} else {
				buffer.WriteRune(focus)
				state = stateLongOption
			}

		// Long option
		case stateLongOption:
			if seperator {
				option = buffer.String()
				buffer.Reset()
				fo := f.FlagSet.Lookup(option)

				if fo == nil {
					if option == "help" {
						return flag.ErrHelp
					}
					return fmt.Errorf("option %q: %w", option, ErrOptionNotFound)
				}

				if trySetBool(fo) {
					state = stateDefault
				} else {
					state = stateValueExpected
				}
			} else if focus == '=' {
				option = buffer.String()
				buffer.Reset()
				state = stateValueExpected
			} else {
				buffer.WriteRune(focus)
			}

		// Single letter options/flags
		case stateShortOptions:
			option = buffer.String()
			buffer.Reset()
			if focus == '=' {
				state = stateValueExpected
			} else {
				fo := f.FlagSet.Lookup(option)
				if fo == nil {
					if option == "h" {
						return flag.ErrHelp
					}
					return fmt.Errorf("option %q: %w", option, ErrOptionNotFound)
				}

				if !trySetBool(fo) {
					if seperator {
						state = stateValueExpected
					} else {
						return fmt.Errorf("option %q: %w", option, ErrOptionNotFlag)
					}
				} else if seperator {
					state = stateDefault
				} else {
					buffer.WriteRune(focus)
				}
			}

		// Option name has been buffered, expecting a value.
		case stateValueExpected:
			if seperator {
				if fo := f.FlagSet.Lookup(option); fo != nil {
					fo.Value.Set(buffer.String())
					buffer.Reset()
					state = stateDefault
				} else {
					return fmt.Errorf("option %q: %w", option, ErrOptionNotFound)
				}
			} else {
				buffer.WriteRune(focus)
			}

		// No more options, arguments only
		case stateArgumentOnly:
			if seperator {
				f.args = append(f.args, buffer.String())
				buffer.Reset()
			} else {
				buffer.WriteRune(focus)
			}
		}
	}

	return nil
}
