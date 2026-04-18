package cards

// allSchemas lists every card schema. SchemaFor() performs the lookup.
var allSchemas = []Schema{
	{
		CardType: TypeTrilogy, CardSubtype: "",
		Sections: []Section{
			{"## Series Arc", true},
			{"## Book-by-Book Progression", true},
			{"## Series-Level Antagonist Ladder", true},
			{"## Series Thematic Questions", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "Overview",
		Sections: []Section{
			{"## Setting", true},
			{"## Tone and Genre", true},
			{"## Core Concept", true},
			{"## Thematic Backdrop", true},
			{"## Reader Experience", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "History",
		Sections: []Section{
			{"## Timeline", true},
			{"## Key Events", true},
			{"## World-Shaping Conflicts", true},
			{"## Current State", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "Political",
		Sections: []Section{
			{"## Power Structure", true},
			{"## Governing Bodies", true},
			{"## Faction Landscape", true},
			{"## Political Tensions", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "Technology",
		Sections: []Section{
			{"## Technology Level", true},
			{"## Key Technologies", true},
			{"## Social Impact", true},
			{"## Constraints and Limits", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "Culture",
		Sections: []Section{
			{"## Social Norms", true},
			{"## Cultural Values", true},
			{"## Class Structure", true},
			{"## Language Notes", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "Geography",
		Sections: []Section{
			{"## Physical Description", true},
			{"## Climate and Conditions", true},
			{"## Strategic Significance", true},
			{"## Notable Features", true},
		},
	},
	{
		CardType: TypeWorld, CardSubtype: "Constraints",
		Sections: []Section{
			{"## Continuity Rules", true},
			{"## Established Facts", true},
			{"## Series Constraints", true},
		},
	},
	{
		CardType: TypeNovel, CardSubtype: "",
		Sections: []Section{
			{"## Story Summary", true},
			{"## Act Structure", true},
			{"## 40-Chapter Map", true},
			{"## Protagonist Journey", true},
			{"## Central Thematic Question", true},
		},
	},
	{
		CardType: TypeCharacter, CardSubtype: TierFull,
		Sections: []Section{
			{"## Identity", true},
			{"## Want", true},
			{"## Need", true},
			{"## Fear", true},
			{"## Misbelief", true},
			{"## Wound", true},
			{"## Arc Progression", true},
			{"## Psychology", true},
			{"## Voice and Mannerisms", true},
			{"## Skills and Knowledge", true},
			{"## Physical Description", true},
			{"## Hard Limits", true},
			{"## Relationship Dynamics", true},
			{"## Active Chapters", true},
			{"## Voice Anchor", true},
		},
	},
	{
		CardType: TypeCharacter, CardSubtype: TierSketch,
		Sections: []Section{
			{"## Identity", true},
			{"## Wound", true},
			{"## Misbelief", true},
			{"## Archetype Stage", true},
			{"## Thematic Role", true},
			{"## Estimated Full Appearance", true},
		},
	},
	{
		CardType: TypeFaction, CardSubtype: TierFull,
		Sections: []Section{
			{"## Overview", true},
			{"## Leadership", true},
			{"## Goals and Methods", true},
			{"## Internal Tensions", true},
			{"## Relationship to Protagonist", true},
			{"## Relationship to Threats", true},
		},
	},
	{
		CardType: TypeFaction, CardSubtype: TierSketch,
		Sections: []Section{
			{"## Overview", true},
			{"## Leadership", true},
			{"## Goals", true},
			{"## Estimated Full Appearance", true},
		},
	},
	{
		CardType: TypeThreat, CardSubtype: TierFull,
		Sections: []Section{
			{"## Nature and Scope", true},
			{"## Origin", true},
			{"## Methods", true},
			{"## Escalation Arc", true},
			{"## Connection to Antagonist", true},
		},
	},
	{
		CardType: TypeThreat, CardSubtype: TierSketch,
		Sections: []Section{
			{"## Overview", true},
			{"## Nature", true},
			{"## Estimated Full Appearance", true},
		},
	},
	{
		CardType: TypeChapter, CardSubtype: "",
		Sections: []Section{
			{"## Chapter Type", true},
			{"## Story Circle Beats", true},
			{"## POV Character", true},
			{"## Scene Beats", true},
			{"## Subplot Threads", false},
		},
	},
	{
		CardType: TypeScene, CardSubtype: "",
		Sections: []Section{
			{"## Chapter Reference", true},
			{"## Story Function", true},
			{"## Opening Hook", true},
			{"## Turn Point", true},
			{"## Closing Beat", true},
			{"## Characters Present", true},
		},
	},
}

// SchemaFor returns the Schema for the given card type/subtype, or nil if not found.
func SchemaFor(cardType, subtype string) *Schema {
	for i := range allSchemas {
		if allSchemas[i].CardType == cardType && allSchemas[i].CardSubtype == subtype {
			return &allSchemas[i]
		}
	}
	return nil
}

// RequiredSectionsList formats the required sections as a Markdown bullet list.
func RequiredSectionsList(cardType, subtype string) string {
	schema := SchemaFor(cardType, subtype)
	if schema == nil {
		return ""
	}
	var out string
	for _, sec := range schema.Sections {
		if sec.Required {
			out += "  - " + sec.Heading + "\n"
		}
	}
	return out
}
