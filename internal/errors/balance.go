package errors

import "fmt"

type NoWithdrawalsError struct {
}

type LowBalanceError struct {
	CurrentBalance int
}

func (w *NoWithdrawalsError) Error() string {
	return "no withdrawals for current user"
}

func (b LowBalanceError) Error() string {
	return fmt.Sprintf("not enought balance, now: %d", b.CurrentBalance)
}
