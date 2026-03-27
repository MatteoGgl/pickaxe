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
