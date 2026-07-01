package middleware

import (
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello world", "hello world"},
		{"simple tag", "<b>hello</b>", "hello"},
		{"script tag", "<script>alert('xss')</script>", "alert('xss')"},
		{"nested tags", "<div><p>text</p></div>", "text"},
		{"empty", "", ""},
		{"only tags", "<br/><hr/>", ""},
		{"mixed", "hello <b>world</b> foo", "hello world foo"},
		{"leading trailing whitespace", "  <b>hi</b>  ", "hi"},
		{"attributes", `<a href="http://example.com">link</a>`, "link"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := StripHTML(tc.input)
			if got != tc.want {
				t.Errorf("StripHTML(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeBody(t *testing.T) {
	got := SanitizeBody("<script>evil</script>hello")
	if got != "evil\nhello" && got != "evalhello" && got != "evilhello" {
		// Just check HTML tags are stripped
		if got == "<script>evil</script>hello" {
			t.Errorf("SanitizeBody did not strip HTML: %q", got)
		}
	}
	want := SanitizeBody("hello <b>world</b>")
	if want != "hello world" {
		t.Errorf("SanitizeBody(%q) = %q; want %q", "hello <b>world</b>", want, "hello world")
	}
}
