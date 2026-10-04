// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/internal/cli"
	"os"
)

func main() { os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr)) }
