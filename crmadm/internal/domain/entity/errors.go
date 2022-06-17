package entity

import "fmt"

type ErrForbiden struct {
	Role string
	Msg  string
}

func (e *ErrForbiden) Error() string {
	return fmt.Sprintf("This action is resticted for role: %s - %s", e.Role, e.Msg)
}

type ErrValidation struct {
	Msg string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error: %s", e.Msg)
}
