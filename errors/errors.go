package errors

import "fmt"

// ErrorSource represents the source of an error
type ErrorSource int

const (
	UnknownSource ErrorSource = iota
	LocalNode
	RemoteNode
	Network
	Database
)

// String returns a string representation of the ErrorSource
func (es ErrorSource) String() string {
	switch es {
	case LocalNode:
		return "Local Node"
	case RemoteNode:
		return "Remote Node"
	case Network:
		return "Network"
	case Database:
		return "Database"
	default:
		return "Unknown Source"
	}
}

type E struct {
	// Error message that give more detail context
	Msg string

	// Represent is underlying error
	Err error

	// Show whether the error is something that temporary and we can retry again.
	Recoverable bool

	// represent the source of the error. Who is the owner. Who make the mistake to rise the error
	Source ErrorSource
}

// NewE creates a new E error. It takes a message, an underlying error, and a recoverable flag.
// Any number of them can be added. it is not necessary to add all of them.
func NewE(params ...interface{}) error {
	e := E{}
	for _, p := range params {
		switch p.(type) {
		case string:
			e.Msg = p.(string)
		case error:
			e.Err = p.(error)
		case bool:
			e.Recoverable = p.(bool)
		}
	}

	return e
}

// Error returns the string representation of the error
func (e E) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s. Error: %s", e.Msg, e.Err.Error())
	}
	return e.Msg
}

// Unwrap returns the underlying error, making E compatible with errors.Unwrap
func (e E) Unwrap() error {
	return e.Err
}
