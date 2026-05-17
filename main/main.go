package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/carapace-sh/revset"
)

func main() {
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
