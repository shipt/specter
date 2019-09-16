package cmd

import (
	"github.com/namsral/flag"
)

// IsFlagPassed checked to see if the flag is passed, so we can determine if its passed with the same value as default.
func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
