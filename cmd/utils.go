package cmd

import (
	"errors"
	"fmt"
	"os"

	sdtdclient "github.com/thelande/sdtd_client/pkg/sdtd_client"
)

// Checks if the given error indicates that Alloc's Server Fixes are not
// installed, notifies the user if this is the case, and then exits with code 1.
// Returns if the given error is of a different nature.
func CheckAllocsMissing(err error) {
	if errors.Is(err, sdtdclient.ErrAllocsModNotInstalled) {
		fmt.Fprintln(
			os.Stderr,
			"This command requires Alloc's Server Fixes to be installed on the server.",
		)
		os.Exit(1)
	} else {
		fmt.Println("nope")
	}
}
