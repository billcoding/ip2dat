package main

import (
	"fmt"
	"os"

	"github.com/billcoding/ip2dat/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
