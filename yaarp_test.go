package yaarp

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
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
	ffs.SetOutput(ioutil.Discard)
	yfs := &FlagSet{FlagSet: ffs}
	if err := yfs.Parse([]string{"-h", "help!"}); err != flag.ErrHelp {
		t.Error("Expected to get the help error. Did not.")
	}
}

// TestHelpLong will test that the help dialog shows up with --help
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

	line, err := outbuf.ReadString('\x00')
	if err != io.EOF {
		t.Fatalf("Expected no error when reading from help buffer, got: %s", err)
	}
	if line != "Usage of test:\n  -a\tis for apple\n" {
		t.Errorf("Expected to get a specific output from help output, got: %q", line)
	}
}

func stringEquality(t *testing.T, expected, gotten string) {
	t.Helper()
	if expected != gotten {
		t.Errorf("Expcted %q, got %q", expected, gotten)
	}
}
