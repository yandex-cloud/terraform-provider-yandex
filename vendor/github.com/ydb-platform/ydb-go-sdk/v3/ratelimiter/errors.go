package ratelimiter

type AcquireError interface {
	error

	Amount() uint64
	Unwrap() error
}
