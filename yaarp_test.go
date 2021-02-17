package yaarp

import (
	"flag"
	"testing"
)

// // TestArgsToArray is a general test of the FlagSet.argsToArray function
// func TestArgsToArray(t *testing.T) {
// 	t.Parallel()

// }

// TestGeneral will test basic functionlity
func TestGeneral(t *testing.T) {
	ffs := flag.NewFlagSet("test", flag.PanicOnError)
	storyval := ffs.String("story", "for", "the purpose of the story")
	tubaval := ffs.String("tuba", "brump", "this is a tuba")
	hval := ffs.String("h", "q", "what about the story")
	tflag := ffs.Bool("t", false, "what about the story")
	aflag := ffs.Bool("a", false, "what's this story of?")
	bflag := ffs.Bool("b", false, "bees")
	isflag := ffs.Bool("is", false, "what is this?")
	yfs := &FlagSet{FlagSet:ffs}
	yfs.Parse([]string{"-th=is", "--is", "the", "--story", "of", "-a", "girl"})

	if yfs.NArg() != 2 {
		t.Fatalf("Expected to have 2 argments, got %d", yfs.NArg())
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

func stringEquality(t *testing.T, expected, gotten string) {
	t.Helper()
	if expected != gotten {
		t.Errorf("Expcted %q, got %q", expected, gotten)
	}
}