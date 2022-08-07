package errors

import "fmt"

type UserAlreadyExistsError struct {
	User string
}

func (err *UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user %s already exists", err.User)
}

type InvalidUserError struct{}

func (err *InvalidUserError) Error() string {
	return fmt.Sprint("invalid username or password")
}
