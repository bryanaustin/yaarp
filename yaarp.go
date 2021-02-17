package yaarp

import (
	"flag"
	"os"
	"strings"
	// "fmt"
)

// FlagSet is the yaarp wrapper for parinsg flags.
type FlagSet struct {
	*flag.FlagSet
	parsed bool
	args   []string
}

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

// VisitAll not implmented. Does anyone use it?

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	var state, i1, i2 int
	// i2 := -1
	var option string
	buffer := &strings.Builder{}

	trySetBool := func() (valueset bool) {
		if fo := f.FlagSet.Lookup(option); fo != nil {
			if bv, ok := fo.Value.(BoolFlagValue); ok && bv.IsBoolFlag() {
				bv.Set("true")
				valueset = true
			}
		} else {
			//ERROR
		}
		return
	}

	for ; ; i2++ {
		var seperator bool
		var focus rune

		if i2 == len(arguments[i1]) {
			seperator = true
		} else {
			if i2 > len(arguments[i1]) {
				i1++
				if i1 >= len(arguments) {
					break
				}
				i2 = 0
			}
			focus = []rune(arguments[i1])[i2]
		}

		// if seperator {
		// 	fmt.Println(fmt.Sprintf("%d seperator", state))
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

		// Reciving an argument (not an option), keep buffering
		// until we come across a seperator
		case stateBufArgument:
			if seperator {
				f.args = append(f.args, buffer.String())
				buffer.Reset()
				state = stateDefault // argument captured, return to defualt state
			} else {
				buffer.WriteRune(focus)
			}

		// This will probably be an option. There are other situations were it
		// not be, but it probably is.
		case stateOptionStart:
			if seperator {
				f.args = append(f.args, string(focus))
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
				if trySetBool() {
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
				if !trySetBool() {
					if seperator {
						state = stateValueExpected
					} else {
						// ERROR case
					}
				} else if seperator {
					state = stateDefault
				} else {
					buffer.WriteRune(focus)
				}
			}

		// Option name has been buffered, expecting a vlaue.
		case stateValueExpected:
			if seperator {
				if fo := f.FlagSet.Lookup(option); fo != nil {
					fo.Value.Set(buffer.String())
					buffer.Reset()
					state = stateDefault
				} else {
					// ERROR
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
