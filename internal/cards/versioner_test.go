package cards

import (
	"strings"
	"testing"
)

func TestLock_ContainsStatusLocked(t *testing.T) {
	content := "## Setting\nA frontier world.\n"
	locked := Lock(content, TypeWorld, "Overview", 1)

	if !strings.Contains(locked, "**Status:** LOCKED") {
		t.Error("locked card missing '**Status:** LOCKED'")
	}
	if !strings.Contains(locked, "**Version:** v1") {
		t.Error("locked card missing '**Version:** v1'")
	}
	if !strings.Contains(locked, "## Setting") {
		t.Error("locked card missing original content section")
	}
}

func TestIsLocked(t *testing.T) {
	locked := Lock("## Setting\nContent.\n", TypeWorld, "Overview", 1)
	if !IsLocked(locked) {
		t.Error("IsLocked() returned false for a locked card")
	}
	if IsLocked("## Setting\nContent.\n") {
		t.Error("IsLocked() returned true for an unlocked draft")
	}
}

func TestExtractBody(t *testing.T) {
	content := "## Setting\nContent.\n"
	locked := Lock(content, TypeWorld, "Overview", 1)
	body := ExtractBody(locked)

	if !strings.Contains(body, "## Setting") {
		t.Error("ExtractBody() missing original section heading")
	}
	if strings.Contains(body, "**Status:** LOCKED") {
		t.Error("ExtractBody() should not contain the lock header")
	}
}
