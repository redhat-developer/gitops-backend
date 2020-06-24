package httpapi

import (
	"testing"

	"github.com/rhd-gitops-examples/gitops-backend/test"
)

var _ DriverIdentifier = (*URLDriverIdentifier)(nil)

func TestIdentify(t *testing.T) {
	urlTests := []struct {
		gitURL  string
		want    string
		wantErr string
	}{
		{"https://github.com/myorg/myrepo.git", "github", ""},
		{"https://gitlab.com/myorg/myrepo/myother.git", "gitlab", ""},
		{"https://scm.example.com/myorg/myother.git", "", "unable to identify driver"},
	}

	identifier := NewDriverIdentifier()
	for _, tt := range urlTests {
		t.Run(tt.gitURL, func(rt *testing.T) {
			driver, err := identifier.Identify(tt.gitURL)
			if !test.MatchError(rt, tt.wantErr, err) {
				rt.Errorf("error failed to match, got %#v, want %s", err, tt.wantErr)
			}

			if driver != tt.want {
				rt.Errorf("got %s, want %s", driver, tt.want)
			}
		})
	}
}
