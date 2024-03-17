package model

import "fmt"

type (
	CliError struct {
		Reason string
	}
	DiffError struct {
		Reason string
	}
)

func (e CliError) Error() string {
	return fmt.Sprintf("cli error: %s", e.Reason)
}

func (e DiffError) Error() string {
	return fmt.Sprintf("diff error: %s", e.Reason)
}
