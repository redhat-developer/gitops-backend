package git

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"sigs.k8s.io/yaml"
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

func (u *URLDriverIdentifier) AddDriversFromFile(filename string) error {
	d := map[string]string{}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	err = yaml.Unmarshal(b, &d)
	if err != nil {
		return fmt.Errorf("failed to parse file %q: %w", filename, err)
	}
	for k, v := range d {
		u.hosts[k] = v
	}
	return nil
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
