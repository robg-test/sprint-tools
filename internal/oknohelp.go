package internal

import (
	"log"
	"time"
)

type OkNoHelpSession struct {
	ID        string
	CreatedAt time.Time
}

var (
	oknohelpSessions = map[string]*OkNoHelpSession{}
)

func CreateOkNoHelpSession() *OkNoHelpSession {
	s := &OkNoHelpSession{
		ID:        newSessionID(),
		CreatedAt: time.Now(),
	}

	if DB != nil {
		_, err := DB.Exec(
			`INSERT INTO oknohelp_sessions (id, created_at) VALUES (?, ?)`,
			s.ID, s.CreatedAt.Unix(),
		)
		if err != nil {
			log.Printf("oknohelp session insert failed, falling back to memory: %s", err)
			storyPointingMu.Lock()
			oknohelpSessions[s.ID] = s
			storyPointingMu.Unlock()
		}
		return s
	}

	storyPointingMu.Lock()
	oknohelpSessions[s.ID] = s
	storyPointingMu.Unlock()
	return s
}

func GetOkNoHelpSession(id string) (*OkNoHelpSession, bool) {
	if DB != nil {
		var createdUnix int64
		row := DB.QueryRow(
			`SELECT created_at FROM oknohelp_sessions WHERE id = ?`,
			id,
		)
		if err := row.Scan(&createdUnix); err == nil {
			return &OkNoHelpSession{
				ID:        id,
				CreatedAt: time.Unix(createdUnix, 0),
			}, true
		}
	}

	storyPointingMu.RLock()
	s, ok := oknohelpSessions[id]
	storyPointingMu.RUnlock()
	return s, ok
}
