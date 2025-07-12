package hosts

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrUnknown      = Error("failed to get certificate")
	ErrConn         = Error("connection error")
	ErrConnTimedout = Error("connection timed out")
	ErrConnRefused  = Error("connection refused")
	ErrCertInvalid  = Error("invalid certificate")
)
