package commands

import (
	"testing"

	"github.com/CogappLabs/linear-cli/internal/client"
)

func TestThreadRoot(t *testing.T) {
	reply := client.Comment{ID: "aa11-reply"}
	reply.Parent = &struct {
		ID string `json:"id"`
	}{ID: "d1b98624-root"}
	comments := []client.Comment{
		{ID: "d1b98624-root"},
		{ID: "d1c00000-other"},
		reply,
	}

	cases := []struct {
		name, ref, want string
		wantErr         bool
	}{
		{"url prefix finds the comment", "d1b98624", "d1b98624-root", false},
		{"comment- anchor is accepted", "comment-d1b98624", "d1b98624-root", false},
		{"full id", "d1c00000-other", "d1c00000-other", false},
		{"a reply resolves to its thread's root", "aa11", "d1b98624-root", false},
		{"ambiguous prefix", "d1", "", true},
		{"no match", "ffff", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := threadRoot(comments, tc.ref)
			if tc.wantErr != (err != nil) {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
