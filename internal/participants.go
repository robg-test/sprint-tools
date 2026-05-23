package internal

import (
	"sort"
	"sync"
	"time"
)

type ParticipantStore struct {
	mu       sync.RWMutex
	bySess   map[string]map[string]participantEntry
	dbTable  string
	hasDBVal func() bool
}

type participantEntry struct {
	role   string
	joined time.Time
}

func newParticipantStore(table string) *ParticipantStore {
	return &ParticipantStore{
		bySess:   map[string]map[string]participantEntry{},
		dbTable:  table,
		hasDBVal: func() bool { return DB != nil },
	}
}

func (p *ParticipantStore) Remove(sessionID, name string) {
	if p.hasDBVal() {
		_, _ = DB.Exec(
			`DELETE FROM `+p.dbTable+` WHERE session_id = ? AND name = ?`,
			sessionID, name,
		)
		return
	}
	p.mu.Lock()
	if m, ok := p.bySess[sessionID]; ok {
		delete(m, name)
	}
	p.mu.Unlock()
}

func (p *ParticipantStore) Add(sessionID, name string) {
	p.AddWithRole(sessionID, name, "play")
}

func (p *ParticipantStore) AddWithRole(sessionID, name, role string) {
	now := time.Now()
	if p.hasDBVal() {
		_, _ = DB.Exec(
			`INSERT INTO `+p.dbTable+` (session_id, name, joined_at, role)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT(session_id, name) DO UPDATE SET role=excluded.role`,
			sessionID, name, now.Unix(), role,
		)
		return
	}
	p.mu.Lock()
	if _, ok := p.bySess[sessionID]; !ok {
		p.bySess[sessionID] = map[string]participantEntry{}
	}
	if existing, exists := p.bySess[sessionID][name]; exists {
		existing.role = role
		p.bySess[sessionID][name] = existing
	} else {
		p.bySess[sessionID][name] = participantEntry{role: role, joined: now}
	}
	p.mu.Unlock()
}

func (p *ParticipantStore) List(sessionID string) []string {
	return p.listFiltered(sessionID, "")
}

func (p *ParticipantStore) Players(sessionID string) []string {
	return p.listFiltered(sessionID, "play")
}

func (p *ParticipantStore) Watchers(sessionID string) []string {
	return p.listFiltered(sessionID, "watch")
}

func (p *ParticipantStore) listFiltered(sessionID, roleFilter string) []string {
	if p.hasDBVal() {
		var query string
		var args []any
		if roleFilter == "" {
			query = `SELECT name FROM ` + p.dbTable + ` WHERE session_id = ? ORDER BY joined_at ASC`
			args = []any{sessionID}
		} else {
			query = `SELECT name FROM ` + p.dbTable + ` WHERE session_id = ? AND role = ? ORDER BY joined_at ASC`
			args = []any{sessionID, roleFilter}
		}
		rows, err := DB.Query(query, args...)
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
	for n, e := range m {
		if roleFilter != "" && e.role != roleFilter {
			continue
		}
		entries = append(entries, entry{n, e.joined})
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
