package errors

import "fmt"

type NoWithdrawalsError struct {
}

type LowBalanceError struct {
	CurrentBalance float32
}

func (w *NoWithdrawalsError) Error() string {
	return "no withdrawals for current user"
}

func (b LowBalanceError) Error() string {
	return fmt.Sprintf("not enought balance, now: %f", b.CurrentBalance)
}
