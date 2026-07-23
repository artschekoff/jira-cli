package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCommentID(t *testing.T) {
	cases := []struct {
		name, raw, want string
		wantErr         bool
	}{
		{"object", `{"id":"10042","body":"hi"}`, "10042", false},
		{"array", `[{"id":"10043"}]`, "10043", false},
		{"whitespace", "  {\"id\":\"10044\"}\n", "10044", false},
		{"missing", `{"body":"hi"}`, "", true},
		{"garbage", `not json`, "", true},
		{"empty array", `[]`, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseCommentID(c.raw)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, c.wantErr)
			}
			if got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestValidateADF(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	good := write("good.json", `{"version":1,"type":"doc","content":[]}`)
	if err := validateADF(good); err != nil {
		t.Fatalf("valid ADF rejected: %v", err)
	}

	for _, c := range []struct{ name, content string }{
		{"not-doc.json", `{"version":1,"type":"paragraph"}`},
		{"no-version.json", `{"type":"doc"}`},
		{"bad-json.json", `{oops`},
	} {
		if err := validateADF(write(c.name, c.content)); err == nil {
			t.Fatalf("%s: expected error, got nil", c.name)
		}
	}

	if err := validateADF(filepath.Join(dir, "nope.json")); err == nil {
		t.Fatal("missing file: expected error, got nil")
	}
}
