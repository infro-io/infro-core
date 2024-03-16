package model

import "fmt"

type (
	NoChangesError struct {
	}
	CliError struct {
		Reason string
	}
	DiffError struct {
		Reason string
	}
)

func (e NoChangesError) Error() string {
	return "no changes"
}

func (e CliError) Error() string {
	return fmt.Sprintf("cli error: %s", e.Reason)
}

func (e DiffError) Error() string {
	return fmt.Sprintf("diff error: %s", e.Reason)
}
