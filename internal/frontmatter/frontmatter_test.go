package frontmatter_test

import (
	"testing"

	"github.com/matteoggl/pickaxe/internal/frontmatter"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no frontmatter",
			input: "# Hello",
			want:  "# Hello",
		},
		{
			name:  "basic",
			input: "---\ntitle: foo\n---\n\n# Hello",
			want:  "\n# Hello",
		},
		{
			name:  "empty frontmatter",
			input: "---\n---\n",
			want:  "",
		},
		{
			name:  "frontmatter only",
			input: "---\ntitle: x\n---",
			want:  "",
		},
		{
			name:  "no closing delimiter",
			input: "---\ntitle: x\nno close",
			want:  "---\ntitle: x\nno close",
		},
		{
			name:  "HR later in doc",
			input: "---\ntags: [a]\n---\n# H\n\n---\n\nMore",
			want:  "# H\n\n---\n\nMore",
		},
		{
			name:  "windows CRLF",
			input: "---\r\ntitle: x\r\n---\r\n\r\n# Hello",
			want:  "\r\n# Hello",
		},
		{
			name:  "leading BOM",
			input: "\xEF\xBB\xBF---\ntitle: x\n---\n# Hello",
			want:  "# Hello",
		},
		{
			name:  "BOM no frontmatter",
			input: "\xEF\xBB\xBF# Hello",
			want:  "# Hello",
		},
		{
			name:  "dashes inside content",
			input: "---\nx: 1\n---\ntext --- here",
			want:  "text --- here",
		},
		{
			name:  "not at byte 0",
			input: " ---\ntitle: x\n---\n",
			want:  " ---\ntitle: x\n---\n",
		},
		{
			name:  "empty file",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := frontmatter.Strip(tc.input)
			if got != tc.want {
				t.Errorf("Strip(%q)\n got  %q\n want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   string
		wantBody string
	}{
		{
			name:     "basic",
			input:    "---\ntitle: foo\n---\n\n# Hello",
			wantFM:   "---\ntitle: foo\n---\n",
			wantBody: "\n# Hello",
		},
		{
			name:     "no frontmatter",
			input:    "# Hello",
			wantFM:   "",
			wantBody: "# Hello",
		},
		{
			name:     "BOM with frontmatter",
			input:    "\xEF\xBB\xBF---\ntitle: x\n---\n# Hello",
			wantFM:   "---\ntitle: x\n---\n",
			wantBody: "# Hello",
		},
		{
			name:     "BOM no frontmatter",
			input:    "\xEF\xBB\xBF# Hello",
			wantFM:   "",
			wantBody: "# Hello",
		},
		{
			name:     "CRLF",
			input:    "---\r\ntitle: x\r\n---\r\n\r\n# Hello",
			wantFM:   "---\r\ntitle: x\r\n---\r\n",
			wantBody: "\r\n# Hello",
		},
		{
			name:     "no closing delimiter",
			input:    "---\ntitle: x\nno close",
			wantFM:   "",
			wantBody: "---\ntitle: x\nno close",
		},
		{
			name:     "empty frontmatter",
			input:    "---\n---\n",
			wantFM:   "---\n---\n",
			wantBody: "",
		},
		{
			name:     "frontmatter only no trailing newline",
			input:    "---\ntitle: x\n---",
			wantFM:   "---\ntitle: x\n---",
			wantBody: "",
		},
		{
			name:     "empty file",
			input:    "",
			wantFM:   "",
			wantBody: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fm, body := frontmatter.Extract(tc.input)
			if fm != tc.wantFM {
				t.Errorf("Extract(%q) fm\n got  %q\n want %q", tc.input, fm, tc.wantFM)
			}
			if body != tc.wantBody {
				t.Errorf("Extract(%q) body\n got  %q\n want %q", tc.input, body, tc.wantBody)
			}
		})
	}
}

func TestExtract_ConsistentWithStrip(t *testing.T) {
	cases := []string{
		"# Hello",
		"---\ntitle: foo\n---\n\n# Hello",
		"---\n---\n",
		"---\ntitle: x\n---",
		"---\ntitle: x\nno close",
		"---\ntags: [a]\n---\n# H\n\n---\n\nMore",
		"---\r\ntitle: x\r\n---\r\n\r\n# Hello",
		"\xEF\xBB\xBF---\ntitle: x\n---\n# Hello",
		"\xEF\xBB\xBF# Hello",
		"---\nx: 1\n---\ntext --- here",
		" ---\ntitle: x\n---\n",
		"",
	}
	for _, input := range cases {
		_, body := frontmatter.Extract(input)
		stripped := frontmatter.Strip(input)
		if body != stripped {
			t.Errorf("Extract body != Strip for input %q\n Extract: %q\n Strip:   %q", input, body, stripped)
		}
	}
}
