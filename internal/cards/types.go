package cards

// Card type constants — mirror queue constants for interoperability.
const (
	TypeWorld       = "WORLD"
	TypeCharacter   = "CHARACTER"
	TypeFaction     = "FACTION"
	TypeThreat      = "THREAT"
	TypeTrilogy     = "TRILOGY"
	TypeNovel       = "NOVEL"
	TypeChapter     = "CHAPTER"
	TypeScene       = "SCENE"
	TypeConstraints = "CONSTRAINTS"
)

const (
	TierFull   = "FULL"
	TierSketch = "SKETCH"
)

const (
	StatusDraft  = "DRAFT"
	StatusLocked = "LOCKED"
)

// Failure is a single card validation failure returned by Validate.
type Failure struct {
	ID          string
	Description string
	Section     string // which section failed, if applicable
}

// Section is a required section within a card identified by its ## heading.
type Section struct {
	Heading  string
	Required bool
}

// Schema defines required sections for a card type/subtype combination.
type Schema struct {
	CardType    string
	CardSubtype string
	Sections    []Section
}
