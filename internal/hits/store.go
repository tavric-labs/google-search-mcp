// Package hits keeps short-lived search hit metadata so later tools can open or
// pack evidence by opaque hit id without re-sending full URLs in every prompt.
package hits

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Hit is one slim search result.
type Hit struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Display string `json:"display_link,omitempty"`
	Query   string `json:"query,omitempty"`
	Weak    bool   `json:"weak,omitempty"`
}

type entry struct {
	hit Hit
	exp time.Time
}

// Store is an in-memory TTL map of hits.
type Store struct {
	mu   sync.Mutex
	ttl  time.Duration
	data map[string]entry
}

// New returns a store that forgets hits after ttl.
func New(ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &Store{ttl: ttl, data: map[string]entry{}}
}

// Put stores hits and assigns stable ids derived from the URL.
func (s *Store) Put(hits []Hit) []Hit {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked()
	out := make([]Hit, 0, len(hits))
	now := time.Now()
	for _, hit := range hits {
		id := idFor(hit.URL)
		hit.ID = id
		s.data[id] = entry{hit: hit, exp: now.Add(s.ttl)}
		out = append(out, hit)
	}
	return out
}

// Get returns a hit by id.
func (s *Store) Get(id string) (Hit, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked()
	entry, ok := s.data[id]
	if !ok {
		return Hit{}, false
	}
	return entry.hit, true
}

// GetMany returns hits for the given ids, skipping missing ones.
func (s *Store) GetMany(ids []string) []Hit {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked()
	out := make([]Hit, 0, len(ids))
	for _, id := range ids {
		if entry, ok := s.data[id]; ok {
			out = append(out, entry.hit)
		}
	}
	return out
}

func (s *Store) gcLocked() {
	now := time.Now()
	for id, entry := range s.data {
		if now.After(entry.exp) {
			delete(s.data, id)
		}
	}
}

func idFor(url string) string {
	sum := sha256.Sum256([]byte(url))
	return "h_" + hex.EncodeToString(sum[:8])
}
