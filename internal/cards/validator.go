package cards

import (
	"fmt"
	"strings"
)

// Validate checks card content against its schema and returns any failures.
// An empty failure list means the card is content-valid.
func Validate(cardType, subtype, content string) []Failure {
	var failures []Failure

	// Check for placeholder text anywhere in the content.
	lower := strings.ToLower(content)
	if strings.Contains(lower, "[tbd]") ||
		strings.Contains(lower, "[todo]") ||
		strings.Contains(lower, "placeholder") {
		failures = append(failures, Failure{
			ID:          "2.3",
			Description: "Card contains placeholder text ([TBD], [TODO], or 'placeholder')",
		})
	}

	schema := SchemaFor(cardType, subtype)
	if schema == nil {
		return failures
	}

	for _, sec := range schema.Sections {
		if !sec.Required {
			continue
		}
		if !sectionPresent(content, sec.Heading) {
			failures = append(failures, Failure{
				ID:          "2.2",
				Description: fmt.Sprintf("Required section missing: %s", sec.Heading),
				Section:     sec.Heading,
			})
		} else if !sectionHasContent(content, sec.Heading) {
			failures = append(failures, Failure{
				ID:          "2.2",
				Description: fmt.Sprintf("Required section is empty: %s", sec.Heading),
				Section:     sec.Heading,
			})
		}
	}

	return failures
}

// sectionPresent reports whether heading appears as a line in content.
func sectionPresent(content, heading string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == heading {
			return true
		}
	}
	return false
}

// sectionHasContent reports whether the section has at least one non-blank line
// of text after its heading and before the next heading.
func sectionHasContent(content, heading string) bool {
	lines := strings.Split(content, "\n")
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == heading {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "# ") {
				return false // reached next section with no content
			}
			if trimmed != "" {
				return true
			}
		}
	}
	return false
}
