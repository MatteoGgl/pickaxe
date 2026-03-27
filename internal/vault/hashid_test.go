package vault_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/matteoggl/pickaxe/internal/vault"
)

func TestHashEntry_Deterministic(t *testing.T) {
	h1 := vault.HashEntry("my-notes")
	h2 := vault.HashEntry("my-notes")
	if h1 != h2 {
		t.Errorf("hash not deterministic: %q vs %q", h1, h2)
	}
}

func TestHashEntry_DifferentNames(t *testing.T) {
	h1 := vault.HashEntry("alpha")
	h2 := vault.HashEntry("beta")
	if h1 == h2 {
		t.Errorf("different names produced same hash: %q", h1)
	}
}

func TestHashEntry_Length(t *testing.T) {
	h := vault.HashEntry("anything")
	if len(h) != 16 {
		t.Errorf("expected 16 hex chars, got %d: %q", len(h), h)
	}
}

func TestNewHashTable_SingleEntry_PrefixLen2(t *testing.T) {
	entries := []vault.Entry{{Name: "only-one"}}
	ht := vault.NewHashTable(entries)
	if ht.ShortPrefixLen("only-one") != 2 {
		t.Errorf("single entry: expected prefix len 2, got %d", ht.ShortPrefixLen("only-one"))
	}
}

func TestNewHashTable_NoPrefixCollision(t *testing.T) {
	// Two entries whose hashes differ at position 2+
	entries := []vault.Entry{{Name: "alpha"}, {Name: "beta"}}
	ht := vault.NewHashTable(entries)
	// Both should get prefix len >= 2 and their prefixes must be distinct
	pa := ht.FullHash("alpha")[:ht.ShortPrefixLen("alpha")]
	pb := ht.FullHash("beta")[:ht.ShortPrefixLen("beta")]
	if pa == pb {
		t.Errorf("prefix collision: both got %q", pa)
	}
}

func TestNewHashTable_ShortHash_8Chars(t *testing.T) {
	entries := []vault.Entry{{Name: "my-notes"}}
	ht := vault.NewHashTable(entries)
	if len(ht.ShortHash("my-notes")) != 8 {
		t.Errorf("expected 8-char short hash, got %d", len(ht.ShortHash("my-notes")))
	}
}

func TestResolvePrefix_ExactHashPrefix(t *testing.T) {
	entries := []vault.Entry{{Name: "alpha"}, {Name: "beta"}}
	ht := vault.NewHashTable(entries)
	hash := ht.FullHash("alpha")
	// Use 4-char prefix to be safe
	name, err := ht.ResolvePrefix(hash[:4])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "alpha" {
		t.Errorf("expected %q, got %q", "alpha", name)
	}
}

func TestResolvePrefix_NotFound(t *testing.T) {
	entries := []vault.Entry{{Name: "alpha"}}
	ht := vault.NewHashTable(entries)
	_, err := ht.ResolvePrefix("0000000000000000")
	if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestResolvePrefix_Ambiguous(t *testing.T) {
	// Build a table with a manufactured collision at prefix length 2.
	// We need two entries whose hashes share the first 2 chars.
	// Brute-force find such a pair.
	var names []string
	seen := map[string]string{} // prefix2 -> name
	for i := 0; len(names) < 2; i++ {
		name := string(rune('a'+i%26)) + string(rune('a'+i/26))
		h := vault.HashEntry(name)
		p := h[:2]
		if existing, ok := seen[p]; ok {
			names = []string{existing, name}
			break
		}
		seen[p] = name
	}

	entries := []vault.Entry{{Name: names[0]}, {Name: names[1]}}
	ht := vault.NewHashTable(entries)

	// Use the shared 2-char prefix
	sharedPrefix := vault.HashEntry(names[0])[:2]
	_, err := ht.ResolvePrefix(sharedPrefix)
	if !errors.Is(err, vault.ErrAmbiguousHash) {
		t.Errorf("expected ErrAmbiguousHash, got %v", err)
	}
}

func TestResolvePrefix_CaseInsensitive(t *testing.T) {
	entries := []vault.Entry{{Name: "alpha"}}
	ht := vault.NewHashTable(entries)
	hash := ht.FullHash("alpha")
	upper := strings.ToUpper(hash[:4])
	name, err := ht.ResolvePrefix(upper)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "alpha" {
		t.Errorf("expected alpha, got %q", name)
	}
}
