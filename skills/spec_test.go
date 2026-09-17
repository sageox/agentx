package skills

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifest(t *testing.T) {
	t.Run("minimal skill", func(t *testing.T) {
		m, err := ParseManifest([]byte("---\nname: pdf-processing\ndescription: Extract PDF text.\n---\n\nBody.\n"))
		require.NoError(t, err)
		assert.Equal(t, "pdf-processing", m.Name)
		assert.Equal(t, "Extract PDF text.", m.Description)
		require.NoError(t, m.Validate("pdf-processing"))
	})

	t.Run("every optional field round-trips", func(t *testing.T) {
		m, err := ParseManifest([]byte(`---
name: code-review
description: Reviews code.
license: Apache-2.0
compatibility: Requires git and jq
metadata:
  author: example-org
  version: "1.0"
allowed-tools: Bash(git:*) Read
---
Body.
`))
		require.NoError(t, err)
		assert.Equal(t, "Apache-2.0", m.License)
		assert.Equal(t, "Requires git and jq", m.Compatibility)
		assert.Equal(t, map[string]string{"author": "example-org", "version": "1.0"}, m.Metadata)
		assert.Equal(t, "Bash(git:*) Read", m.AllowedTools)
	})

	// The spec is young and still gaining optional fields. Rejecting a key we have
	// not learned yet would reject skills that are valid for a newer agent.
	t.Run("unknown keys are ignored, not rejected", func(t *testing.T) {
		m, err := ParseManifest([]byte("---\nname: x\ndescription: d\nsome-future-field: hello\n---\n"))
		require.NoError(t, err)
		assert.Equal(t, "x", m.Name)
	})

	t.Run("no frontmatter", func(t *testing.T) {
		_, err := ParseManifest([]byte("# Just a heading\n\nBody.\n"))
		assert.ErrorIs(t, err, ErrNoFrontmatter)
	})

	// A --- later in the document is a horizontal rule. Treating it as an opening
	// fence would parse a skill's prose as metadata.
	t.Run("horizontal rule is not frontmatter", func(t *testing.T) {
		_, err := ParseManifest([]byte("Intro paragraph.\n\n---\n\nname: not-really\n"))
		assert.ErrorIs(t, err, ErrNoFrontmatter)
	})

	t.Run("unterminated frontmatter is absent, not whole-document metadata", func(t *testing.T) {
		_, err := ParseManifest([]byte("---\nname: x\ndescription: d\n\nBody with no closing fence.\n"))
		assert.ErrorIs(t, err, ErrNoFrontmatter)
	})

	t.Run("CRLF line endings", func(t *testing.T) {
		m, err := ParseManifest([]byte("---\r\nname: windows-skill\r\ndescription: d\r\n---\r\nBody.\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "windows-skill", m.Name)
	})

	t.Run("malformed YAML reports a parse error", func(t *testing.T) {
		_, err := ParseManifest([]byte("---\nname: [unclosed\n---\n"))
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNoFrontmatter)
	})
}

func TestManifestValidate(t *testing.T) {
	valid := Manifest{Name: "data-analysis", Description: "Analyzes data."}

	tests := []struct {
		name     string
		manifest Manifest
		dirName  string
		wantErr  string
	}{
		{"valid", valid, "data-analysis", ""},
		{"valid without the directory check", valid, "", ""},
		{"missing name", Manifest{Description: "d"}, "", "name is required"},
		{"missing description", Manifest{Name: "x"}, "", "description is required"},
		{"blank description", Manifest{Name: "x", Description: "   "}, "", "description is required"},
		{"uppercase name", Manifest{Name: "PDF-Processing", Description: "d"}, "", "lowercase letters"},
		{"underscore in name", Manifest{Name: "pdf_processing", Description: "d"}, "", "lowercase letters"},
		{"leading hyphen", Manifest{Name: "-pdf", Description: "d"}, "", "start or end with a hyphen"},
		{"trailing hyphen", Manifest{Name: "pdf-", Description: "d"}, "", "start or end with a hyphen"},
		{"consecutive hyphens", Manifest{Name: "pdf--processing", Description: "d"}, "", "consecutive hyphens"},
		{"name too long", Manifest{Name: strings.Repeat("a", MaxNameLen+1), Description: "d"}, "", "limit is 64"},
		{"description too long", Manifest{Name: "x", Description: strings.Repeat("d", MaxDescriptionLen+1)}, "", "limit is 1024"},
		{"compatibility too long", Manifest{Name: "x", Description: "d", Compatibility: strings.Repeat("c", MaxCompatibilityLen+1)}, "", "limit is 500"},
		{"name does not match directory", valid, "something-else", `must match its directory name`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate(tt.dirName)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}

	// An author fixing a skill wants the whole list in one pass; the round trip
	// through an agent to surface violations one at a time is the expensive part.
	t.Run("reports every violation at once", func(t *testing.T) {
		err := Manifest{Name: "-Bad--Name-", Compatibility: strings.Repeat("c", MaxCompatibilityLen+1)}.Validate("other")
		require.Error(t, err)
		for _, want := range []string{"start or end with a hyphen", "description is required", "must match its directory name", "limit is 500"} {
			assert.Contains(t, err.Error(), want)
		}
	})

	t.Run("name at exactly the limit is valid", func(t *testing.T) {
		assert.NoError(t, Manifest{Name: strings.Repeat("a", MaxNameLen), Description: "d"}.Validate(""))
	})
}

// TestLimitsCountCharactersNotBytes guards the spec's wording: the limits are
// character limits. Measuring len() would reject a description of exactly 1024
// accented characters at twice its true length, penalizing every manifest not
// written in ASCII.
func TestLimitsCountCharactersNotBytes(t *testing.T) {
	atLimit := Manifest{
		Name:          strings.Repeat("a", MaxNameLen),
		Description:   strings.Repeat("é", MaxDescriptionLen),
		Compatibility: strings.Repeat("é", MaxCompatibilityLen),
	}
	require.Greater(t, len(atLimit.Description), MaxDescriptionLen, "test needs multi-byte runes to be meaningful")
	assert.NoError(t, atLimit.Validate(""))

	overByOne := Manifest{Name: "x", Description: strings.Repeat("é", MaxDescriptionLen+1)}
	require.Error(t, overByOne.Validate(""))
	assert.Contains(t, overByOne.Validate("").Error(), "description is 1025 characters")
}

// TestClosingFenceMustBeExact: a line that merely starts with --- is not a
// closing fence. Accepting one truncates the metadata at an arbitrary point and
// yields a manifest a conforming consumer rejects — while this parser reports
// success.
func TestClosingFenceMustBeExact(t *testing.T) {
	notFences := map[string]string{
		"four hyphens":        "---\nname: x\ndescription: d\n----\n",
		"hyphens plus text":   "---\nname: x\ndescription: d\n---bad\n",
		"indented fence":      "---\nname: x\ndescription: d\n  ---\n",
		"fence with trailing": "---\nname: x\ndescription: d\n--- \n",
	}
	for name, doc := range notFences {
		t.Run(name, func(t *testing.T) {
			_, err := ParseManifest([]byte(doc))
			assert.ErrorIs(t, err, ErrNoFrontmatter)
		})
	}

	t.Run("exact fence closes the block", func(t *testing.T) {
		m, err := ParseManifest([]byte("---\nname: x\ndescription: d\n---\nBody.\n"))
		require.NoError(t, err)
		assert.Equal(t, "x", m.Name)
	})

	// A --- inside the body must not be mistaken for the fence a second time, and
	// a file whose final line is the fence with no trailing newline still parses.
	t.Run("fence as the final line without a trailing newline", func(t *testing.T) {
		m, err := ParseManifest([]byte("---\nname: x\ndescription: d\n---"))
		require.NoError(t, err)
		assert.Equal(t, "x", m.Name)
	})

	t.Run("empty frontmatter parses but fails validation", func(t *testing.T) {
		m, err := ParseManifest([]byte("---\n---\nBody.\n"))
		require.NoError(t, err)
		assert.ErrorContains(t, m.Validate(""), "name is required")
	})
}
