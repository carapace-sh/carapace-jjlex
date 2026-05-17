package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: revset <expression> or revset --complete <cursor> <expression>")
		os.Exit(1)
	}

	if os.Args[1] == "--complete" {
		if len(os.Args) < 4 {
			println("usage: revset --complete <cursor> <expression>")
			os.Exit(1)
		}
		cursor, err := strconv.Atoi(os.Args[2])
		if err != nil {
			println(fmt.Sprintf("invalid cursor: %v", err))
			os.Exit(1)
		}
		ctx := revset.ParseForCompletion(os.Args[3], cursor)
		m, err := json.Marshal(ctx)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		fmt.Println(string(m))
		return
	}

	expression, err := revset.Parse(os.Args[1])
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
