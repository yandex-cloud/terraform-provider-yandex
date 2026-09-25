package xerrors

type alreadyHasTxError struct {
	currentTx string
}

func (err *alreadyHasTxError) Error() string {
	return "сonn already has an open currentTx: " + err.currentTx
}

func (err *alreadyHasTxError) As(target interface{}) bool {
	switch t := target.(type) {
	case *alreadyHasTxError:
		t.currentTx = err.currentTx

		return true
	default:
		return false
	}
}

func AlreadyHasTx(txID string) error {
	return &alreadyHasTxError{currentTx: txID}
}

func IsAlreadyHasTx(err error) bool {
	var txErr *alreadyHasTxError

	return As(err, &txErr)
}
