package internal

import (
	"sync"
	"time"
)

type VoteStore struct {
	mu      sync.RWMutex
	votes   map[string]map[string]string
	dbTable string
}

func newVoteStore(table string) *VoteStore {
	return &VoteStore{
		votes:   map[string]map[string]string{},
		dbTable: table,
	}
}

func (s *VoteStore) Cast(sessionID, name, value string) {
	if DB != nil {
		_, _ = DB.Exec(
			`INSERT INTO `+s.dbTable+` (session_id, name, value, voted_at)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT(session_id, name) DO UPDATE SET
			   value=excluded.value, voted_at=excluded.voted_at`,
			sessionID, name, value, time.Now().Unix(),
		)
		return
	}
	s.mu.Lock()
	if _, ok := s.votes[sessionID]; !ok {
		s.votes[sessionID] = map[string]string{}
	}
	s.votes[sessionID][name] = value
	s.mu.Unlock()
}

func (s *VoteStore) Get(sessionID, name string) string {
	if DB != nil {
		var v string
		row := DB.QueryRow(
			`SELECT value FROM `+s.dbTable+` WHERE session_id = ? AND name = ?`,
			sessionID, name,
		)
		_ = row.Scan(&v)
		return v
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.votes[sessionID][name]
}

func (s *VoteStore) Voters(sessionID string) map[string]bool {
	out := map[string]bool{}
	if DB != nil {
		rows, err := DB.Query(
			`SELECT name FROM `+s.dbTable+` WHERE session_id = ?`,
			sessionID,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var n string
				if err := rows.Scan(&n); err == nil {
					out[n] = true
				}
			}
		}
		return out
	}
	s.mu.RLock()
	for n := range s.votes[sessionID] {
		out[n] = true
	}
	s.mu.RUnlock()
	return out
}

func (s *VoteStore) ClearOne(sessionID, name string) {
	if DB != nil {
		_, _ = DB.Exec(`DELETE FROM `+s.dbTable+` WHERE session_id = ? AND name = ?`, sessionID, name)
		return
	}
	s.mu.Lock()
	if m, ok := s.votes[sessionID]; ok {
		delete(m, name)
	}
	s.mu.Unlock()
}

func (s *VoteStore) All(sessionID string) map[string]string {
	out := map[string]string{}
	if DB != nil {
		rows, err := DB.Query(
			`SELECT name, value FROM `+s.dbTable+` WHERE session_id = ?`,
			sessionID,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var n, v string
				if err := rows.Scan(&n, &v); err == nil {
					out[n] = v
				}
			}
		}
		return out
	}
	s.mu.RLock()
	for n, v := range s.votes[sessionID] {
		out[n] = v
	}
	s.mu.RUnlock()
	return out
}

func (s *VoteStore) Clear(sessionID string) {
	if DB != nil {
		_, _ = DB.Exec(`DELETE FROM `+s.dbTable+` WHERE session_id = ?`, sessionID)
		return
	}
	s.mu.Lock()
	delete(s.votes, sessionID)
	s.mu.Unlock()
}

var (
	StoryPointingVotes = newVoteStore("story_pointing_votes")
	OkNoHelpVotes      = newVoteStore("oknohelp_votes")
)
