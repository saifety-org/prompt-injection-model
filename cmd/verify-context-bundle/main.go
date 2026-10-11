// Command verify-context-bundle checks a local research bundle, without inference.
package main

import (
	"fmt"
	"github.com/saifety-org/prompt-injection-model/bundle"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: verify-context-bundle DIRECTORY")
		os.Exit(2)
	}
	m, e := bundle.Load(os.Args[1])
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Printf("%s %s verified (research only)\n", m.ID, m.Version)
}
