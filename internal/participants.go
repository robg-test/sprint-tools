package internal

import (
	"sort"
	"sync"
	"time"
)

type participantStore struct {
	mu       sync.RWMutex
	bySess   map[string]map[string]time.Time
	dbTable  string
	hasDBVal func() bool
}

func newParticipantStore(table string) *participantStore {
	return &participantStore{
		bySess:   map[string]map[string]time.Time{},
		dbTable:  table,
		hasDBVal: func() bool { return DB != nil },
	}
}

func (p *participantStore) Add(sessionID, name string) {
	now := time.Now()
	if p.hasDBVal() {
		_, _ = DB.Exec(
			`INSERT OR IGNORE INTO `+p.dbTable+` (session_id, name, joined_at) VALUES (?, ?, ?)`,
			sessionID, name, now.Unix(),
		)
		return
	}
	p.mu.Lock()
	if _, ok := p.bySess[sessionID]; !ok {
		p.bySess[sessionID] = map[string]time.Time{}
	}
	if _, exists := p.bySess[sessionID][name]; !exists {
		p.bySess[sessionID][name] = now
	}
	p.mu.Unlock()
}

func (p *participantStore) List(sessionID string) []string {
	if p.hasDBVal() {
		rows, err := DB.Query(
			`SELECT name FROM `+p.dbTable+` WHERE session_id = ? ORDER BY joined_at ASC`,
			sessionID,
		)
		if err != nil {
			return nil
		}
		defer rows.Close()
		var names []string
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err == nil {
				names = append(names, n)
			}
		}
		return names
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	m := p.bySess[sessionID]
	type entry struct {
		name string
		at   time.Time
	}
	entries := make([]entry, 0, len(m))
	for n, t := range m {
		entries = append(entries, entry{n, t})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].at.Before(entries[j].at) })
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.name
	}
	return names
}

var (
	StoryPointingParticipants = newParticipantStore("story_pointing_participants")
	OkNoHelpParticipants      = newParticipantStore("oknohelp_participants")
)
