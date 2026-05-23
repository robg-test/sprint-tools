package internal

import (
	"crypto/rand"
	"encoding/base32"
	"log"
	"strings"
	"sync"
	"time"
)

type StoryPointingSession struct {
	ID        string
	CreatedAt time.Time
}

var (
	storyPointingMu       sync.RWMutex
	storyPointingSessions = map[string]*StoryPointingSession{}
)

func CreateStoryPointingSession() *StoryPointingSession {
	s := &StoryPointingSession{
		ID:        newSessionID(),
		CreatedAt: time.Now(),
	}

	if DB != nil {
		_, err := DB.Exec(
			`INSERT INTO story_pointing_sessions (id, created_at) VALUES (?, ?)`,
			s.ID, s.CreatedAt.Unix(),
		)
		if err != nil {
			log.Printf("story pointing session insert failed, falling back to memory: %s", err)
			storeInMemory(s)
		}
		return s
	}

	storeInMemory(s)
	return s
}

func GetStoryPointingSession(id string) (*StoryPointingSession, bool) {
	if DB != nil {
		var createdUnix int64
		row := DB.QueryRow(
			`SELECT created_at FROM story_pointing_sessions WHERE id = ?`,
			id,
		)
		if err := row.Scan(&createdUnix); err == nil {
			return &StoryPointingSession{
				ID:        id,
				CreatedAt: time.Unix(createdUnix, 0),
			}, true
		}
	}

	storyPointingMu.RLock()
	s, ok := storyPointingSessions[id]
	storyPointingMu.RUnlock()
	return s, ok
}

func storeInMemory(s *StoryPointingSession) {
	storyPointingMu.Lock()
	storyPointingSessions[s.ID] = s
	storyPointingMu.Unlock()
}

func newSessionID() string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	id := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return strings.ToLower(id)
}
