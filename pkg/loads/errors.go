// SPDX-License-Identifier: AGPL-3.0-only
package loads

import "strings"

type DomainError struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (e DomainError) Error() string { return e.Path + ": " + e.Message }

type ValidationErrors []DomainError

func (e ValidationErrors) Error() string {
	var lines []string
	for _, v := range e {
		lines = append(lines, v.Error())
	}
	return strings.Join(lines, "\n")
}
