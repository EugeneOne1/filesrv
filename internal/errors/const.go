package errors

// Error is the [error] that may be constant.
type Error string

func (e Error) Error() string { return string(e) }
