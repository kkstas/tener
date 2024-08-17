package url

import (
	"fmt"
	"testing"
)

func TestBuildURL(t *testing.T) {
	cases := []struct {
		stage string
		parts []string
		want  string
	}{
		{"", []string{"one", "two"}, "/one/two"},
		{"", []string{}, "/"},
		{"stage", []string{}, "/stage"},
		{"stage", []string{"one", "two"}, "/stage/one/two"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("creates %q for parts %v and STAGE_NAME env %q", c.want, c.parts, c.stage), func(t *testing.T) {
			got := buildURL(c.stage, c.parts...)

			if got != c.want {
				t.Errorf("got %q want %q", got, c.want)
			}
		})

	}
}
