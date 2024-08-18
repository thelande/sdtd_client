/*
Copyright © 2024 Tom Helander <thomas.helander@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
