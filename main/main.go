package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/carapace-sh/carapace-jjlex/pkg/fileset"
	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: jjlex <mode> <expression>")
		println("modes: revset, revset-complete, fileset, fileset-complete, fileset-bare, fileset-bare-complete")
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "revset":
		parseRevset(os.Args[2:])
	case "revset-complete":
		completeRevset(os.Args[2:])
	case "fileset":
		parseFileset(os.Args[2:])
	case "fileset-complete":
		completeFileset(os.Args[2:])
	case "fileset-bare":
		parseFilesetBare(os.Args[2:])
	case "fileset-bare-complete":
		completeFilesetBare(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func parseRevset(args []string) {
	if len(args) < 1 {
		println("usage: jjlex revset <expression>")
		os.Exit(1)
	}
	expression, err := revset.Parse(args[0])
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	m, err := json.Marshal(expression)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(m))
}

func completeRevset(args []string) {
	if len(args) < 2 {
		println("usage: jjlex revset-complete <cursor> <expression>")
		os.Exit(1)
	}
	cursor, err := strconv.Atoi(args[0])
	if err != nil {
		println(fmt.Sprintf("invalid cursor: %v", err))
		os.Exit(1)
	}
	ctx := revset.ParseForCompletion(args[1], cursor)
	m, err := json.Marshal(ctx)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(m))
}

func parseFileset(args []string) {
	if len(args) < 1 {
		println("usage: jjlex fileset <expression>")
		os.Exit(1)
	}
	expression, err := fileset.Parse(args[0])
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	m, err := json.Marshal(expression)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(m))
}

func completeFileset(args []string) {
	if len(args) < 2 {
		println("usage: jjlex fileset-complete <cursor> <expression>")
		os.Exit(1)
	}
	cursor, err := strconv.Atoi(args[0])
	if err != nil {
		println(fmt.Sprintf("invalid cursor: %v", err))
		os.Exit(1)
	}
	ctx := fileset.ParseForCompletion(args[1], cursor)
	m, err := json.Marshal(ctx)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(m))
}

func parseFilesetBare(args []string) {
	if len(args) < 1 {
		println("usage: jjlex fileset-bare <expression>")
		os.Exit(1)
	}
	expression, err := fileset.ParseProgramOrBareString(args[0])
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	m, err := json.Marshal(expression)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(m))
}

func completeFilesetBare(args []string) {
	// Same as fileset-complete for now - bare string completion
	// is handled at a higher level by the caller
	completeFileset(args)
}