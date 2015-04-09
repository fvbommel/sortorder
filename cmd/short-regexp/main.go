package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fvbommel/util"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	// Remove trailing empty line if present
	if N := len(lines); N > 0 && lines[N-1] == "" {
		lines = lines[:N-1]
	}
	os.Stdout.WriteString(util.ShortRegexpString(lines...))
}
