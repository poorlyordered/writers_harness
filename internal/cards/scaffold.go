package cards

import (
	"fmt"
	"strings"
)

const scaffoldPlaceholder = "[Write content here]"

// GenerateScaffold returns an empty card template with all required ## headings.
// Writers fill in the content before Phase 3 runs; Phase 3 loads and refines it.
func GenerateScaffold(cardType, cardSubtype, name string) string {
	schema := SchemaFor(cardType, cardSubtype)

	label := cardType
	if cardSubtype != "" {
		label += " — " + cardSubtype
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s — %s\n\n", label, name))
	sb.WriteString("**Status:** TEMPLATE\n\n")
	sb.WriteString("---\n\n")

	if schema == nil {
		sb.WriteString(scaffoldPlaceholder + "\n")
		return sb.String()
	}

	for _, section := range schema.Sections {
		sb.WriteString(section.Heading + "\n\n")
		sb.WriteString(scaffoldPlaceholder + "\n\n")
	}
	return strings.TrimRight(sb.String(), "\n") + "\n"
}

// IsScaffoldOnly reports whether the content is an unfilled scaffold
// (every section still has the placeholder — writer has not started).
func IsScaffoldOnly(content string) bool {
	return strings.Contains(content, scaffoldPlaceholder)
}
