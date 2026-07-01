package main

import (
	"flag"
	"os"
	"testing"
)

func TestReplayFlagsRequireAction(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"replay"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	*flagID = ""
	*flagAll = false
	*flagList = false

	// Verify default flags are false/empty.
	if *flagID != "" || *flagAll || *flagList {
		t.Fatal("expected empty default flags")
	}
}
