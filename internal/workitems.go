package internal

import (
	"sync"
	"time"
)

type WorkItem struct {
	SessionID string
	Summary   string
	UpdatedAt time.Time
}

type WorkItemStore struct {
	mu      sync.RWMutex
	items   map[string]*WorkItem
	dbTable string
}

func newWorkItemStore(table string) *WorkItemStore {
	return &WorkItemStore{
		items:   map[string]*WorkItem{},
		dbTable: table,
	}
}

func (s *WorkItemStore) Get(sessionID string) *WorkItem {
	if DB != nil {
		row := DB.QueryRow(
			`SELECT summary, updated_at FROM `+s.dbTable+` WHERE session_id = ?`,
			sessionID,
		)
		var w WorkItem
		var updated int64
		if err := row.Scan(&w.Summary, &updated); err == nil {
			w.SessionID = sessionID
			w.UpdatedAt = time.Unix(updated, 0)
			return &w
		}
		return &WorkItem{SessionID: sessionID}
	}
	s.mu.RLock()
	w, ok := s.items[sessionID]
	s.mu.RUnlock()
	if !ok {
		return &WorkItem{SessionID: sessionID}
	}
	return w
}

func (s *WorkItemStore) Set(w *WorkItem) {
	w.UpdatedAt = time.Now()
	if DB != nil {
		_, _ = DB.Exec(
			`INSERT INTO `+s.dbTable+` (session_id, summary, updated_at)
			 VALUES (?, ?, ?)
			 ON CONFLICT(session_id) DO UPDATE SET
			   summary=excluded.summary, updated_at=excluded.updated_at`,
			w.SessionID, w.Summary, w.UpdatedAt.Unix(),
		)
		return
	}
	s.mu.Lock()
	s.items[w.SessionID] = w
	s.mu.Unlock()
}

var (
	StoryPointingWorkItems = newWorkItemStore("story_pointing_summaries")
	OkNoHelpWorkItems      = newWorkItemStore("oknohelp_summaries")
)
