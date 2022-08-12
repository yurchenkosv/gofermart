package errors

import "fmt"

type OrderAlreadyAcceptedCurrentUserError struct {
	OrderNumber string
	UserID      int
}

type OrderAlreadyAcceptedDifferentUserError struct {
	OrderNumber string
	UserID      int
}

type OrderFormatError struct {
	OrderNumber string
}

type NoOrdersError struct {
}

type OrderNoChangeError struct {
}

func (err *OrderAlreadyAcceptedCurrentUserError) Error() string {
	return fmt.Sprintf("order with number %s already accepted from user %d", err.OrderNumber, err.UserID)
}

func (err *OrderAlreadyAcceptedDifferentUserError) Error() string {
	return fmt.Sprintf("order with number %s already accepted from different user %d", err.OrderNumber, err.UserID)
}

func (err *OrderFormatError) Error() string {
	return fmt.Sprintf("order %s invalid by format", err.OrderNumber)
}

func (err *NoOrdersError) Error() string {
	return "no orders was found"
}

func (o *OrderNoChangeError) Error() string {
	return "order no change"
}
