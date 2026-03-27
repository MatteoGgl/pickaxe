package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashEntry computes a deterministic 16-char hex ID for an entry name.
func HashEntry(name string) string {
	h := sha256.Sum256([]byte(name))
	return hex.EncodeToString(h[:8])
}

// HashTable maps entry names to their hashes and shortest unique prefix lengths.
type HashTable struct {
	hashes    map[string]string
	prefixLen map[string]int
}

// NewHashTable builds a HashTable from a slice of entries.
func NewHashTable(entries []Entry) *HashTable {
	ht := &HashTable{
		hashes:    make(map[string]string, len(entries)),
		prefixLen: make(map[string]int, len(entries)),
	}
	for _, e := range entries {
		ht.hashes[e.Name] = HashEntry(e.Name)
	}
	// Compute shortest unique prefix per entry (min 2 chars).
	for name, hash := range ht.hashes {
		plen := 2
		for plen < len(hash) {
			prefix := hash[:plen]
			unique := true
			for otherName, otherHash := range ht.hashes {
				if otherName != name && strings.HasPrefix(otherHash, prefix) {
					unique = false
					break
				}
			}
			if unique {
				break
			}
			plen++
		}
		ht.prefixLen[name] = plen
	}
	return ht
}

// FullHash returns the full 16-char hex hash for the named entry.
func (ht *HashTable) FullHash(name string) string {
	return ht.hashes[name]
}

// ShortHash returns the 8-char display hash for the named entry.
func (ht *HashTable) ShortHash(name string) string {
	h := ht.hashes[name]
	if len(h) > 8 {
		return h[:8]
	}
	return h
}

// ShortPrefixLen returns the length of the shortest unique prefix for the named entry,
// capped at 8 (the display hash length).
func (ht *HashTable) ShortPrefixLen(name string) int {
	plen := ht.prefixLen[name]
	if plen > 8 {
		return 8
	}
	return plen
}

// ResolvePrefix finds the entry name whose hash starts with prefix.
// Returns ErrAmbiguousHash if multiple entries match, ErrNotFound if none.
func (ht *HashTable) ResolvePrefix(prefix string) (string, error) {
	prefix = strings.ToLower(prefix)
	var match string
	for name, hash := range ht.hashes {
		if strings.HasPrefix(hash, prefix) {
			if match != "" {
				return "", ErrAmbiguousHash
			}
			match = name
		}
	}
	if match == "" {
		return "", ErrNotFound
	}
	return match, nil
}
