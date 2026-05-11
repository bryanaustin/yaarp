package yaarp

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

// TestGeneral will test basic functionlity
func TestGeneral(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	storyval := ffs.String("story", "for", "the purpose of the story")
	tubaval := ffs.String("tuba", "brump", "this is a tuba")
	hval := ffs.String("h", "q", "what about the story")
	tflag := ffs.Bool("t", false, "what about the story")
	aflag := ffs.Bool("a", false, "what's this story of?")
	bflag := ffs.Bool("b", false, "bees")
	isflag := ffs.Bool("is", false, "what is this?")
	yfs := &FlagSet{FlagSet: ffs}

	if err := yfs.Parse([]string{"-th=is", "--is", "the", "--story", "of", "-a", "girl"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}

	if yfs.NArg() != 2 {
		t.Fatalf("Expected to have 2 arguments, got %d", yfs.NArg())
	}

	stringEquality(t, "the", yfs.Arg(0))
	stringEquality(t, "girl", yfs.Arg(1))
	stringEquality(t, "of", *storyval)
	stringEquality(t, "brump", *tubaval)
	stringEquality(t, "is", *hval)

	if !*tflag {
		t.Error("Expected tflag to be true, it wasn't")
	}

	if !*aflag {
		t.Error("Expected aflag to be true, it wasn't")
	}

	if *bflag {
		t.Error("Expected bflag to be false, it wasn't")
	}

	if !*isflag {
		t.Error("Expected isflag to be true, it wasn't")
	}
}

// TestDashes will test - and -- arguments
func TestDashes(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	outval := ffs.String("o", "q", "output")
	yfs := &FlagSet{FlagSet: ffs}

	yfs.Parse([]string{"-", "-o", "-", "--", "-o", "-"})

	if yfs.NArg() != 3 {
		t.Fatalf("Expected to have 3 arguments, got %d", yfs.NArg())
	}

	stringEquality(t, "-", yfs.Arg(0))
	stringEquality(t, "-o", yfs.Arg(1))
	stringEquality(t, "-", yfs.Arg(2))
	stringEquality(t, "-", *outval)
}

// TestHelpSingle will test that the help dialog shows up with -h
func TestHelpSingle(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.SetOutput(io.Discard)
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"-h", "help!"}); err != flag.ErrHelp {
		t.Error("Expected to get the help error. Did not.")
	}
}

// TestHelpLong will test that the help dialog shows up with --help and
// renders in the new GNU-style default format.
func TestHelpLong(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "is for apple")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Error("Expected to get the help error. Did not.")
	}

	expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  -a" + strings.Repeat(" ", 10) + "is for apple\n" +
		"  -h, --help  show this help and exit\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected help output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestUnicode will test may claims about unicode options
func TestUnicode(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	smiles := ffs.Bool("☺", false, "☺")
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{"-☺"})

	if !*smiles {
		t.Fatal("Expected to have ☺, got ☹")
	}
}

// TestNoArgs will test that everything doesn't break if no args are passed
func TestNoArgs(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{})

	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestEmptyArg will test that everything doesn't break if no args are passed
func TestEmptyArg(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{""})

	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestQuotesInArg will test an option with a value that has quotes
func TestQuotesInArg(t *testing.T) {
	t.Parallel()
	expected := `{"name":{"first":"bob"}}`
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	gotten := ffs.String("json", "{}", "json")
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{"--json", expected})
	stringEquality(t, expected, *gotten)

	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestEmptyNonFirst will test what happens when there is an empty non-first arg
func TestEmptyNonFirst(t *testing.T) {
	t.Parallel()
	expected := "%"
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	gotten := ffs.String("val", "", "value")
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{"--val", expected, ""})
	stringEquality(t, expected, *gotten)

	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestEmptyNonFirstNotLast will test what happens when there is something after a non-first empty arg
func TestEmptyNonFirstNotLast(t *testing.T) {
	t.Parallel()
	expected := "%"
	expectedArg := "words"
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	gotten := ffs.String("val", "", "value")
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{"--val", expected, "", expectedArg})
	stringEquality(t, expected, *gotten)

	if yfs.NArg() != 1 {
		t.Fatalf("Expected to have 1 argument, got %d", yfs.NArg())
	}
	if yfs.Arg(0) != expectedArg {
		t.Errorf("Expected argument %q, got: %q", expectedArg, yfs.Arg(0))
	}
}

func stringEquality(t *testing.T, expected, gotten string) {
	t.Helper()
	if expected != gotten {
		t.Errorf("Expcted %q, got %q", expected, gotten)
	}
}

// TestClusterBoolValueNextToken tests that a clustered short flag like -tv with t bool
// and v non-bool consumes the next token as v's value.
func TestClusterBoolValueNextToken(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	tflag := ffs.Bool("t", false, "t")
	vval := ffs.String("v", "", "v")
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{"-tv", "value"})

	if !*tflag {
		t.Error("Expected tflag to be true, it wasn't")
	}
	stringEquality(t, "value", *vval)
	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestClusterAllBool tests that a clustered short flag of all bool flags sets each true.
func TestClusterAllBool(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	tflag := ffs.Bool("t", false, "t")
	aflag := ffs.Bool("a", false, "a")
	bflag := ffs.Bool("b", false, "b")
	yfs := &FlagSet{FlagSet: ffs}
	yfs.Parse([]string{"-tab"})

	if !*tflag {
		t.Error("Expected tflag to be true, it wasn't")
	}
	if !*aflag {
		t.Error("Expected aflag to be true, it wasn't")
	}
	if !*bflag {
		t.Error("Expected bflag to be true, it wasn't")
	}
	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestErrOptionNotFlag tests that a clustered short flag with a non-bool option
// followed by more characters in the same token returns ErrOptionNotFlag.
func TestErrOptionNotFlag(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "a")
	ffs.String("v", "", "v")
	yfs := &FlagSet{FlagSet: ffs}
	err := yfs.Parse([]string{"-avx"})
	if !errors.Is(err, ErrOptionNotFlag) {
		t.Errorf("Expected error to wrap ErrOptionNotFlag, got: %v", err)
	}
}

// TestUnknownLongOption tests that an undefined long option returns ErrOptionNotFound.
func TestUnknownLongOption(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	yfs := &FlagSet{FlagSet: ffs}
	err := yfs.Parse([]string{"--unknown"})
	if !errors.Is(err, ErrOptionNotFound) {
		t.Errorf("Expected error to wrap ErrOptionNotFound, got: %v", err)
	}
}

// TestUnknownShortOption tests that an undefined short option (other than -h) returns ErrOptionNotFound.
func TestUnknownShortOption(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	yfs := &FlagSet{FlagSet: ffs}
	err := yfs.Parse([]string{"-x"})
	if !errors.Is(err, ErrOptionNotFound) {
		t.Errorf("Expected error to wrap ErrOptionNotFound, got: %v", err)
	}
}

// TestShortHelpDefinedAsFlag tests that when -h is defined as a real flag it is not intercepted as help.
func TestShortHelpDefinedAsFlag(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	hflag := ffs.Bool("h", false, "help")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"-h"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	if !*hflag {
		t.Error("Expected hflag to be true, it wasn't")
	}
}

// TestLongHelpDefinedAsFlag tests that when --help is defined as a real flag it is not intercepted as help.
func TestLongHelpDefinedAsFlag(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	helpflag := ffs.Bool("help", false, "help")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"--help"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	if !*helpflag {
		t.Error("Expected helpflag to be true, it wasn't")
	}
}

// TestDoubleDashAtEnd tests that a trailing -- with no following args terminates cleanly.
func TestDoubleDashAtEnd(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	flagval := ffs.String("flag", "", "flag")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"--flag", "value", "--"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	stringEquality(t, "value", *flagval)
	if yfs.NArg() != 0 {
		t.Fatalf("Expected to have 0 arguments, got %d", yfs.NArg())
	}
}

// TestDefaultUsageMixed exercises the default usage with bool/string/int/backtick
// flags, asserting synopsis, alignment, placeholders, and default-value formatting.
func TestDefaultUsageMixed(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Int("count", 5, "how many")
	ffs.String("file", "main.go", "read from `FILE`")
	ffs.String("output", "out.txt", "where to write")
	ffs.Bool("v", false, "verbose mode")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}

	expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  --count <int>" + strings.Repeat(" ", 6) + "how many (default 5)\n" +
		"  --file <FILE>" + strings.Repeat(" ", 6) + "read from FILE (default \"main.go\")\n" +
		"  --output <string>  where to write (default \"out.txt\")\n" +
		"  -v" + strings.Repeat(" ", 17) + "verbose mode\n" +
		"  -h, --help" + strings.Repeat(" ", 9) + "show this help and exit\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected help output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestDefaultUsageUserHelpSuppressesSynthetic confirms the synthetic
// "-h, --help" row is not appended when the user defines either flag.
func TestDefaultUsageUserHelpSuppressesSynthetic(t *testing.T) {
	t.Parallel()

	t.Run("user-defined -h", func(t *testing.T) {
		outbuf := new(bytes.Buffer)
		ffs := flag.NewFlagSet("test", flag.ContinueOnError)
		ffs.Bool("h", false, "user h flag")
		yfs := &FlagSet{FlagSet: ffs}
		ffs.SetOutput(outbuf)

		if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
			t.Fatalf("Expected ErrHelp, got: %v", err)
		}

		expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
			"\n" +
			"Options:\n" +
			"  -h  user h flag\n"
		if got := outbuf.String(); got != expected {
			t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
		}
		if strings.Contains(outbuf.String(), "--help") {
			t.Errorf("synthetic --help row should be suppressed; got: %q", outbuf.String())
		}
	})

	t.Run("user-defined --help", func(t *testing.T) {
		outbuf := new(bytes.Buffer)
		ffs := flag.NewFlagSet("test", flag.ContinueOnError)
		ffs.Bool("help", false, "user help flag")
		yfs := &FlagSet{FlagSet: ffs}
		ffs.SetOutput(outbuf)

		if err := yfs.Parse([]string{"-h"}); err != flag.ErrHelp {
			t.Fatalf("Expected ErrHelp, got: %v", err)
		}

		expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
			"\n" +
			"Options:\n" +
			"  --help  user help flag\n"
		if got := outbuf.String(); got != expected {
			t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
		}
	})
}

// TestDefaultUsageNoName confirms the synopsis omits the program name when
// the FlagSet has an empty Name.
func TestDefaultUsageNoName(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("", flag.ContinueOnError)
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}

	expected := "Usage: [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  -h, --help  show this help and exit\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestDefaultUsageNoFlags confirms the synthetic help row appears even when
// the FlagSet has no user-defined flags.
func TestDefaultUsageNoFlags(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}

	expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  -h, --help  show this help and exit\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestDefaultUsageFoldsLongSyntax confirms that a syntax cell exceeding the
// fold threshold puts its description on the next line, while short rows
// remain aligned to the description column anchored on the widest
// short-row syntax.
func TestDefaultUsageFoldsLongSyntax(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "the a flag")
	ffs.Bool("b", false, "the b flag")
	ffs.String("really-long-config-name", "default.cfg", "load config")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}

	expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  -a" + strings.Repeat(" ", 10) + "the a flag\n" +
		"  -b" + strings.Repeat(" ", 10) + "the b flag\n" +
		"  --really-long-config-name <string>\n" +
		strings.Repeat(" ", 14) + "load config (default \"default.cfg\")\n" +
		"  -h, --help  show this help and exit\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestDefaultUsageAllFolded confirms that when every row exceeds the fold
// threshold, folded descriptions land at column 4 (indent + 0 + gutter). It
// suppresses the synthetic row by defining both -h and --help, and uses
// backtick-named placeholders to force those user flags to fold as well.
func TestDefaultUsageAllFolded(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("tool", flag.ContinueOnError)
	ffs.String("first-extremely-long-flag", "", "first one")
	ffs.Int("second-extremely-long-flag", 10, "second one")
	ffs.Bool("h", false, "user h `LONG_PLACEHOLDER_TO_FORCE_FOLDING_OF_H`")
	ffs.Bool("help", false, "user help `LONG_PLACEHOLDER_TO_FORCE_FOLDING_OF_HELP`")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)
	yfs.printDefaultUsage()

	expected := "Usage: tool [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  --first-extremely-long-flag <string>\n" +
		"    first one\n" +
		"  -h <LONG_PLACEHOLDER_TO_FORCE_FOLDING_OF_H>\n" +
		"    user h LONG_PLACEHOLDER_TO_FORCE_FOLDING_OF_H\n" +
		"  --help <LONG_PLACEHOLDER_TO_FORCE_FOLDING_OF_HELP>\n" +
		"    user help LONG_PLACEHOLDER_TO_FORCE_FOLDING_OF_HELP\n" +
		"  --second-extremely-long-flag <int>\n" +
		"    second one (default 10)\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestDefaultUsageUnicode confirms width is measured in runes, not bytes:
// a single-rune multi-byte flag name aligns with a single-byte one.
func TestDefaultUsageUnicode(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "is for apple")
	ffs.Bool("☺", false, "smiley")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}

	expected := "Usage: test [OPTIONS] [ARGUMENTS]\n" +
		"\n" +
		"Options:\n" +
		"  -a" + strings.Repeat(" ", 10) + "is for apple\n" +
		"  -☺" + strings.Repeat(" ", 10) + "smiley\n" +
		"  -h, --help  show this help and exit\n"
	if got := outbuf.String(); got != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", got, expected)
	}
}

// TestDefaultUsageEmptyDescription confirms a flag with an empty usage string
// produces a syntax-only row with no trailing whitespace.
func TestDefaultUsageEmptyDescription(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("x", false, "")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}

	out := outbuf.String()
	if !strings.Contains(out, "  -x\n") {
		t.Errorf("expected empty-description row %q in output, got: %q", "  -x\n", out)
	}
	if strings.Contains(out, "  -x ") {
		t.Errorf("empty-description row should have no trailing whitespace, got: %q", out)
	}
}

// TestDefaultUsageNoTabs is a regression guard: yaarp must never emit tab
// characters in the default usage output.
func TestDefaultUsageNoTabs(t *testing.T) {
	t.Parallel()
	outbuf := new(bytes.Buffer)
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "is for apple")
	ffs.Int("count", 5, "how many")
	ffs.String("name", "world", "your name")
	yfs := &FlagSet{FlagSet: ffs}
	ffs.SetOutput(outbuf)

	if err := yfs.Parse([]string{"--help"}); err != flag.ErrHelp {
		t.Fatalf("Expected ErrHelp, got: %v", err)
	}
	if strings.ContainsRune(outbuf.String(), '\t') {
		t.Errorf("output should contain no tab characters; got: %q", outbuf.String())
	}
}

// collectVisited drives a Visit-style iterator and returns the names in the order yielded.
func collectVisited(visit func(func(*flag.Flag))) []string {
	var names []string
	visit(func(fo *flag.Flag) { names = append(names, fo.Name) })
	return names
}

// TestVisitAllOrder confirms VisitAll yields every defined flag in lexicographical order.
func TestVisitAllOrder(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("c", false, "c")
	ffs.Bool("a", false, "a")
	ffs.Bool("b", false, "b")
	yfs := &FlagSet{FlagSet: ffs}
	if got := strings.Join(collectVisited(yfs.VisitAll), ","); got != "a,b,c" {
		t.Errorf("VisitAll order: got %q, want %q", got, "a,b,c")
	}
}

// TestVisitOnlySetFlags confirms Visit yields only flags that have been set during
// parsing, in lexicographical order.
func TestVisitOnlySetFlags(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "a")
	ffs.String("b", "", "b")
	ffs.Bool("c", false, "c")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"-a", "--b=x"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	if got := strings.Join(collectVisited(yfs.Visit), ","); got != "a,b" {
		t.Errorf("Visit names: got %q, want %q", got, "a,b")
	}
}

// TestVisitClusteredBool confirms each bool in a cluster like -tab is routed through
// the FlagSet's Set tracking and appears in Visit.
func TestVisitClusteredBool(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("t", false, "t")
	ffs.Bool("a", false, "a")
	ffs.Bool("b", false, "b")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"-tab"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	if got := strings.Join(collectVisited(yfs.Visit), ","); got != "a,b,t" {
		t.Errorf("Visit names: got %q, want %q", got, "a,b,t")
	}
}

// TestVisitConsumesNextToken confirms a flag whose value comes from the next token is
// tracked in Visit, and the visited Flag reports the consumed value via Value.String().
func TestVisitConsumesNextToken(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.String("v", "", "v")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"-v", "value"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	if got := strings.Join(collectVisited(yfs.Visit), ","); got != "v" {
		t.Errorf("Visit names: got %q, want %q", got, "v")
	}
	yfs.Visit(func(fo *flag.Flag) {
		stringEquality(t, "value", fo.Value.String())
	})
}

// TestVisitBeforeParse confirms that before Parse runs, Visit yields nothing while
// VisitAll still yields the defined flags.
func TestVisitBeforeParse(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("a", false, "a")
	yfs := &FlagSet{FlagSet: ffs}
	if got := collectVisited(yfs.Visit); len(got) != 0 {
		t.Errorf("Visit before Parse: got %v, want empty", got)
	}
	if got := strings.Join(collectVisited(yfs.VisitAll), ","); got != "a" {
		t.Errorf("VisitAll before Parse: got %q, want %q", got, "a")
	}
}

// TestVisitSkipsFailedSet confirms that when Value.Set fails (e.g., --flag= with an
// empty value for a bool flag, which fails strconv.ParseBool), the flag is NOT recorded
// as set, matching stdlib semantics. Parse itself still returns nil because Set errors
// are silently dropped.
func TestVisitSkipsFailedSet(t *testing.T) {
	t.Parallel()
	ffs := flag.NewFlagSet("test", flag.ContinueOnError)
	ffs.Bool("flag", false, "flag")
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"--flag="}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}
	if got := collectVisited(yfs.Visit); len(got) != 0 {
		t.Errorf("Visit should be empty when Set failed; got %v", got)
	}
}

// TestPackageLevelVisit confirms the package-level Visit and VisitAll functions
// delegate to CommandLine and yield the same names as the FlagSet methods. Swaps
// CommandLine for an isolated FlagSet and restores it on exit.
func TestPackageLevelVisit(t *testing.T) {
	saved := CommandLine
	defer func() { CommandLine = saved }()

	ffs := flag.NewFlagSet("p", flag.ContinueOnError)
	ffs.Bool("a", false, "a")
	ffs.String("b", "", "b")
	CommandLine = &FlagSet{FlagSet: ffs}
	if err := CommandLine.Parse([]string{"--b=x"}); err != nil {
		t.Fatalf("Expected parse to run without error, got: %s", err)
	}

	if got := strings.Join(collectVisited(Visit), ","); got != "b" {
		t.Errorf("package Visit: got %q, want %q", got, "b")
	}
	if got := strings.Join(collectVisited(VisitAll), ","); got != "a,b" {
		t.Errorf("package VisitAll: got %q, want %q", got, "a,b")
	}
}
