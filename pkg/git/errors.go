package git

import "net/http"

// IsNotFound returns true if the error represents a NotFound response from an
// upstream service.
func IsNotFound(err error) bool {
	e, ok := err.(SCMError)
	return ok && e.Status == http.StatusNotFound
}

type SCMError struct {
	msg    string
	Status int
}

func (s SCMError) Error() string {
	return s.msg
}
