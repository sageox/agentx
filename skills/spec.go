package skills

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// ManifestName is the file that makes a directory a skill.
const ManifestName = "SKILL.md"

// Spec field limits, from the Agent Skills specification
// (https://agentskills.io/specification).
//
// These are CHARACTER limits, not byte limits, and are enforced with
// utf8.RuneCountInString. Measuring bytes would reject a description of 1024 é
// at twice its true length — every non-ASCII manifest penalized for its
// alphabet.
const (
	MaxNameLen          = 64
	MaxDescriptionLen   = 1024
	MaxCompatibilityLen = 500
)

// Manifest is the parsed frontmatter of a SKILL.md.
//
// Unknown frontmatter keys are ignored rather than rejected: the spec is young
// and still gaining optional fields, and a validator that hard-fails on a key it
// has not learned yet would reject skills that are perfectly valid for a newer
// agent.
type Manifest struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`

	// AllowedTools is a space-separated list of pre-approved tools.
	//
	// EXPERIMENTAL per the spec, and support varies by agent — treat its presence
	// as a capability grant worth a human's attention, not as settled metadata.
	AllowedTools string `yaml:"allowed-tools,omitempty"`
}

// ErrNoFrontmatter reports a SKILL.md with no YAML frontmatter block. The spec
// requires one, so this is a malformed skill rather than an empty one.
var ErrNoFrontmatter = errors.New("skills: SKILL.md has no YAML frontmatter")

// ParseManifest reads the frontmatter of a SKILL.md. It does not validate;
// call Validate for that, so a caller can report a parse failure and a spec
// violation differently.
func ParseManifest(data []byte) (Manifest, error) {
	front, ok := frontmatter(data)
	if !ok {
		return Manifest{}, ErrNoFrontmatter
	}
	var m Manifest
	if err := yaml.Unmarshal(front, &m); err != nil {
		return Manifest{}, fmt.Errorf("skills: parse SKILL.md frontmatter: %w", err)
	}
	return m, nil
}

// Validate checks the manifest against the Agent Skills specification.
//
// dirName is the skill's own directory name, which the spec requires `name` to
// match. Pass "" to skip that one check when validating a manifest that is not
// yet on disk.
//
// Every violation is reported, not just the first: an author fixing a skill
// wants the whole list in one pass, and the round-trip through an agent to
// surface them one at a time is the expensive part.
func (m Manifest) Validate(dirName string) error {
	var problems []string

	switch {
	case m.Name == "":
		problems = append(problems, "name is required")
	case utf8.RuneCountInString(m.Name) > MaxNameLen:
		problems = append(problems, fmt.Sprintf("name is %d characters, limit is %d", utf8.RuneCountInString(m.Name), MaxNameLen))
	}
	if m.Name != "" {
		if bad := invalidNameReason(m.Name); bad != "" {
			problems = append(problems, "name "+bad)
		}
		if dirName != "" && m.Name != dirName {
			problems = append(problems, fmt.Sprintf("name %q must match its directory name %q", m.Name, dirName))
		}
	}

	switch {
	case strings.TrimSpace(m.Description) == "":
		problems = append(problems, "description is required")
	case utf8.RuneCountInString(m.Description) > MaxDescriptionLen:
		problems = append(problems, fmt.Sprintf("description is %d characters, limit is %d", utf8.RuneCountInString(m.Description), MaxDescriptionLen))
	}

	if n := utf8.RuneCountInString(m.Compatibility); n > MaxCompatibilityLen {
		problems = append(problems, fmt.Sprintf("compatibility is %d characters, limit is %d", n, MaxCompatibilityLen))
	}

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("skills: invalid SKILL.md: %s", strings.Join(problems, "; "))
}

// invalidNameReason returns why a name violates the spec's character rules, or
// "" when it is valid. Hand-checked rather than regexp'd so the message can name
// the actual problem instead of restating a pattern.
func invalidNameReason(name string) string {
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return "must not start or end with a hyphen"
	}
	if strings.Contains(name, "--") {
		return "must not contain consecutive hyphens"
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return fmt.Sprintf("may only contain lowercase letters, digits and hyphens (found %q)", r)
		}
	}
	return ""
}

// frontmatter extracts the bytes between the leading --- fences.
//
// The opening fence must be the very first line: a --- appearing later in a
// document is a horizontal rule, and treating it as frontmatter would parse a
// skill's prose as metadata.
//
// The closing fence must be a line of exactly ---. Accepting any line that
// merely STARTS with --- would close the block on ---- or on a --- that begins a
// sentence, silently truncating the metadata and handing back a manifest that a
// conforming consumer rejects.
func frontmatter(data []byte) ([]byte, bool) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, false
	}
	body := text[len("---\n"):]
	for offset := 0; offset <= len(body); {
		line := body[offset:]
		end := strings.IndexByte(line, '\n')
		if end >= 0 {
			line = line[:end]
		}
		if line == "---" {
			return []byte(body[:offset]), true
		}
		if end < 0 {
			break
		}
		offset += end + 1
	}
	// Unterminated frontmatter. Reported as absent rather than silently treating
	// the whole document as metadata.
	return nil, false
}
