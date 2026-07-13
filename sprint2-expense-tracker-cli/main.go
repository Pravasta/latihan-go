package main

import (
	"fmt"
	"os"
	"sprint2-expense-tracker-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
