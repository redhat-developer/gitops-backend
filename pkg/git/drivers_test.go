package git

import (
	"testing"

	"github.com/rhd-gitops-examples/gitops-backend/test"
)

var _ DriverIdentifier = (*URLDriverIdentifier)(nil)

func TestIdentify(t *testing.T) {
	urlTests := []struct {
		gitURL   string
		want     string
		readFile string
		wantErr  string
	}{
		{"https://github.com/myorg/myrepo.git", "github", "", ""},
		{"https://gitlab.com/myorg/myrepo/myother.git", "gitlab", "", ""},
		{"https://gl.example.com/myorg/myrepo/myother.git", "gitlab", "testdata/drivers.yaml", ""},
		{"https://scm.example.com/myorg/myother.git", "", "", "unable to identify driver"},
		{"https://u.example.com/myorg/myrepo/myother.git", "", "doesnotexist.yaml", "failed to open"},
	}

	for _, tt := range urlTests {
		t.Run(tt.gitURL, func(rt *testing.T) {
			identifier := NewDriverIdentifier()
			if tt.readFile != "" {
				err := identifier.AddDriversFromFile(tt.readFile)
				if !test.MatchError(rt, tt.wantErr, err) {
					rt.Errorf("got an error reading the drivers file %s: %s", tt.readFile, err)
				}
				return
			}

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
