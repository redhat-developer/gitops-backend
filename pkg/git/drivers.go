package git

import (
	"fmt"
	"net/url"
)

// URLDriverIdentifier is an implementation of the DriverIdentifier interface
// that looks up hosts in a map.
type URLDriverIdentifier struct {
	hosts map[string]string
}

func (u *URLDriverIdentifier) Identify(gitURL string) (string, error) {
	parsed, err := url.Parse(gitURL)
	if err != nil {
	}
	d, ok := u.hosts[parsed.Host]
	if ok {
		return d, nil
	}
	return "", unknownDriverError{url: gitURL}
}

// NewDriverIdentifier creates and returns a new URLDriverIdentifier.
func NewDriverIdentifier() *URLDriverIdentifier {
	return &URLDriverIdentifier{
		hosts: map[string]string{
			"github.com": "github",
			"gitlab.com": "gitlab",
		},
	}
}

type unknownDriverError struct {
	url string
}

func (e unknownDriverError) Error() string {
	return fmt.Sprintf("unable to identify driver from URL: %s", e.url)
}
