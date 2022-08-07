package errors

import "fmt"

type OrderAlreadyAcceptedCurrentUserError struct {
	OrderNumber int
	UserID      int
}

type OrderAlreadyAcceptedDifferentUser struct {
	OrderNumber int
}

type OrderFormatError struct {
	OrderNumber int
}

type NoOrdersDataError struct {
}

func (err *OrderAlreadyAcceptedCurrentUserError) Error() string {
	return fmt.Sprintf("order with number %d already accepted from user %d", err.OrderNumber, err.UserID)
}

func (err *OrderAlreadyAcceptedDifferentUser) Error() string {
	return fmt.Sprintf("order with number %d already accepted from different user", err.OrderNumber)
}

func (err *OrderFormatError) Error() string {
	return fmt.Sprintf("order %d invalid by format", err.OrderNumber)
}

func (err *NoOrdersDataError) Error() string {
	return "no orders was made for current user"
}
