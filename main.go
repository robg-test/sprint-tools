package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/robgtest/sprint-tools/internal"
	"github.com/robgtest/sprint-tools/web/pages"
)

const (
	defaultTheme = "light"
	darkTheme    = "night"
)


func main() {
	log.SetFlags(log.LstdFlags)
	setup()

	router := chi.NewRouter()
	setupStaticHandlers(router)
	setupPageHandlers(router)

	env := os.Getenv("ENV")

	if env == "production" {
		host := ":443"
		certPath := os.Getenv("CERT_PATH")
		keyPath := os.Getenv("KEY_PATH")

		log.Println("Starting secure server on", host)
		err := http.ListenAndServeTLS(host, certPath, keyPath, internal.UserSessionManager.LoadAndSave(router))
		if err != nil {
			log.Printf("secure server failed: %s", err)
		}
		return
	}

	host := ":7000"
	log.Println("Starting development server on", host)
	err := http.ListenAndServe(host, internal.UserSessionManager.LoadAndSave(router))
	if err != nil {
		log.Fatalf("server failed: %s", err)
	}
}

func setupPageHandlers(router *chi.Mux) {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.IndexPage(themeOrDefault(r))).ServeHTTP(w, r)
	})

	router.Post("/story-pointing/sessions", func(w http.ResponseWriter, r *http.Request) {
		s := internal.CreateStoryPointingSession()
		internal.PutMessage("sp:"+s.ID+":host", "1", r)
		http.Redirect(w, r, "/story-pointing/"+s.ID, http.StatusSeeOther)
	})

	router.Get("/story-pointing/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		theme := themeOrDefault(r)
		sessionURL := absoluteURL(r, "/story-pointing/"+id)
		name := internal.GetMessage("sp:"+id+":name", r)
		role := internal.GetMessage("sp:"+id+":role", r)
		if role == "" {
			name = ""
		}
		isHost := internal.GetMessage("sp:"+id+":host", r) == "1"
		players := internal.StoryPointingParticipants.Players(id)
		watchers := internal.StoryPointingParticipants.Watchers(id)
		workItem := internal.StoryPointingWorkItems.Get(id)
		voters := internal.StoryPointingVotes.Voters(id)
		allVotes := internal.StoryPointingVotes.All(id)
		myVote := internal.StoryPointingVotes.Get(id, name)
		canVote := role != "watch" && name != ""
		templ.Handler(pages.StoryPointingSessionPage(theme, sessionURL, id, name, isHost, canVote, players, watchers, voters, allVotes, myVote, workItem)).ServeHTTP(w, r)
	})

	router.Post("/story-pointing/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Redirect(w, r, "/story-pointing/"+id, http.StatusSeeOther)
			return
		}
		if len(name) > 40 {
			name = name[:40]
		}
		role := r.FormValue("role")
		if role != "watch" {
			role = "play"
		}
		internal.PutMessage("sp:"+id+":name", name, r)
		internal.PutMessage("sp:"+id+":role", role, r)
		internal.StoryPointingParticipants.AddWithRole(id, name, role)
		if role == "watch" {
			internal.StoryPointingVotes.ClearOne(id, name)
		}
		http.Redirect(w, r, "/story-pointing/"+id, http.StatusSeeOther)
	})

	router.Get("/story-pointing/{id}/participants", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		players := internal.StoryPointingParticipants.Players(id)
		watchers := internal.StoryPointingParticipants.Watchers(id)
		templ.Handler(pages.ParticipantsList(players, watchers)).ServeHTTP(w, r)
	})

	router.Post("/ok-no/sessions", func(w http.ResponseWriter, r *http.Request) {
		s := internal.CreateOkNoHelpSession()
		internal.PutMessage("on:"+s.ID+":host", "1", r)
		http.Redirect(w, r, "/ok-no/"+s.ID, http.StatusSeeOther)
	})

	router.Get("/ok-no/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		theme := themeOrDefault(r)
		sessionURL := absoluteURL(r, "/ok-no/"+id)
		name := internal.GetMessage("on:"+id+":name", r)
		role := internal.GetMessage("on:"+id+":role", r)
		if role == "" {
			name = ""
		}
		isHost := internal.GetMessage("on:"+id+":host", r) == "1"
		players := internal.OkNoHelpParticipants.Players(id)
		watchers := internal.OkNoHelpParticipants.Watchers(id)
		workItem := internal.OkNoHelpWorkItems.Get(id)
		voters := internal.OkNoHelpVotes.Voters(id)
		allVotes := internal.OkNoHelpVotes.All(id)
		myVote := internal.OkNoHelpVotes.Get(id, name)
		canVote := role != "watch" && name != ""
		templ.Handler(pages.OkNoHelpSessionPage(theme, sessionURL, id, name, isHost, canVote, players, watchers, voters, allVotes, myVote, workItem)).ServeHTTP(w, r)
	})

	router.Post("/ok-no/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Redirect(w, r, "/ok-no/"+id, http.StatusSeeOther)
			return
		}
		if len(name) > 40 {
			name = name[:40]
		}
		role := r.FormValue("role")
		if role != "watch" {
			role = "play"
		}
		internal.PutMessage("on:"+id+":name", name, r)
		internal.PutMessage("on:"+id+":role", role, r)
		internal.OkNoHelpParticipants.AddWithRole(id, name, role)
		if role == "watch" {
			internal.OkNoHelpVotes.ClearOne(id, name)
		}
		http.Redirect(w, r, "/ok-no/"+id, http.StatusSeeOther)
	})

	router.Get("/ok-no/{id}/participants", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		players := internal.OkNoHelpParticipants.Players(id)
		watchers := internal.OkNoHelpParticipants.Watchers(id)
		templ.Handler(pages.ParticipantsList(players, watchers)).ServeHTTP(w, r)
	})

	router.Get("/story-pointing/{id}/work-item", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		isHost := internal.GetMessage("sp:"+id+":host", r) == "1"
		item := internal.StoryPointingWorkItems.Get(id)
		players := internal.StoryPointingParticipants.Players(id)
		watchers := internal.StoryPointingParticipants.Watchers(id)
		voters := internal.StoryPointingVotes.Voters(id)
		allVotes := internal.StoryPointingVotes.All(id)
		name := internal.GetMessage("sp:"+id+":name", r)
		myVote := internal.StoryPointingVotes.Get(id, name)
		canVote := internal.GetMessage("sp:"+id+":role", r) != "watch" && name != ""
		templ.Handler(pages.WorkItemSlot(item, players, watchers, voters, allVotes, myVote, isHost, canVote, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards)).ServeHTTP(w, r)
	})

	router.Post("/story-pointing/{id}/work-item", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		startRound(w, r, id, "sp", internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards)
	})

	router.Post("/story-pointing/{id}/work-item/summary", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		saveSummary(w, r, id, "sp", internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards)
	})

	router.Post("/story-pointing/{id}/work-item/countdown", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		startCountdown(w, r, id, internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards, "sp")
	})

	router.Post("/story-pointing/{id}/work-item/cancel-countdown", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		cancelCountdown(w, r, id, internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards, "sp")
	})

	router.Post("/story-pointing/{id}/work-item/vote", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		castVote(w, r, id, "sp", internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards)
	})

	router.Post("/story-pointing/{id}/work-item/end", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		endRound(w, r, id, "sp", internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards)
	})

	router.Post("/story-pointing/{id}/work-item/repoint", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		repointRound(w, r, id, "sp", internal.StoryPointingWorkItems, internal.StoryPointingParticipants, internal.StoryPointingVotes, "/story-pointing/"+id+"/work-item", pages.StoryPointingCards)
	})

	router.Get("/ok-no/{id}/work-item", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		isHost := internal.GetMessage("on:"+id+":host", r) == "1"
		item := internal.OkNoHelpWorkItems.Get(id)
		players := internal.OkNoHelpParticipants.Players(id)
		watchers := internal.OkNoHelpParticipants.Watchers(id)
		voters := internal.OkNoHelpVotes.Voters(id)
		allVotes := internal.OkNoHelpVotes.All(id)
		name := internal.GetMessage("on:"+id+":name", r)
		myVote := internal.OkNoHelpVotes.Get(id, name)
		canVote := internal.GetMessage("on:"+id+":role", r) != "watch" && name != ""
		templ.Handler(pages.WorkItemSlot(item, players, watchers, voters, allVotes, myVote, isHost, canVote, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards)).ServeHTTP(w, r)
	})

	router.Post("/ok-no/{id}/work-item", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		startRound(w, r, id, "on", internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards)
	})

	router.Post("/ok-no/{id}/work-item/summary", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		saveSummary(w, r, id, "on", internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards)
	})

	router.Post("/ok-no/{id}/work-item/countdown", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		startCountdown(w, r, id, internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards, "on")
	})

	router.Post("/ok-no/{id}/work-item/cancel-countdown", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		cancelCountdown(w, r, id, internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards, "on")
	})

	router.Post("/ok-no/{id}/work-item/vote", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		castVote(w, r, id, "on", internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards)
	})

	router.Post("/ok-no/{id}/work-item/end", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		endRound(w, r, id, "on", internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards)
	})

	router.Post("/ok-no/{id}/work-item/repoint", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":host", r) != "1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		repointRound(w, r, id, "on", internal.OkNoHelpWorkItems, internal.OkNoHelpParticipants, internal.OkNoHelpVotes, "/ok-no/"+id+"/work-item", pages.OkNoHelpCards)
	})

	router.Get("/story-pointing/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("sp:"+id+":role", r) == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		sseStream(w, r, id)
	})

	router.Get("/ok-no/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		if internal.GetMessage("on:"+id+":role", r) == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		sseStream(w, r, id)
	})

	router.Put("/theme", func(w http.ResponseWriter, r *http.Request) {
		if internal.GetMessage("theme", r) != darkTheme {
			internal.PutMessage("theme", darkTheme, r)
		} else {
			internal.PutMessage("theme", defaultTheme, r)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func sseStream(w http.ResponseWriter, r *http.Request, id string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := internal.SessionEvents.Subscribe(id)
	defer internal.SessionEvents.Unsubscribe(id, ch)
	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "data: refresh\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func repointRound(w http.ResponseWriter, r *http.Request, id, namePrefix string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string) {
	item := store.Get(id)
	item.CountdownUntil = time.Time{}
	store.Set(item)
	internal.SessionEvents.Notify(id)
	name := internal.GetMessage(namePrefix+":"+id+":name", r)
	canVote := internal.GetMessage(namePrefix+":"+id+":role", r) != "watch" && name != ""
	templ.Handler(pages.WorkItemSlot(item, parts.Players(id), parts.Watchers(id), votes.Voters(id), votes.All(id), "", true, canVote, baseURL, cards)).ServeHTTP(w, r)
}

func saveSummary(w http.ResponseWriter, r *http.Request, id, namePrefix string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	item := store.Get(id)
	item.Summary = strings.TrimSpace(r.FormValue("summary"))
	store.Set(item)
	internal.SessionEvents.Notify(id)
	templ.Handler(pages.WorkItemForm(item, baseURL, baseURL+"/summary", cards)).ServeHTTP(w, r)
}

func startRound(w http.ResponseWriter, r *http.Request, id, namePrefix string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	item := &internal.WorkItem{
		SessionID: id,
		Summary:   strings.TrimSpace(r.FormValue("summary")),
		Active:    true,
	}
	store.Set(item)
	votes.Clear(id)
	internal.SessionEvents.Notify(id)
	name := internal.GetMessage(namePrefix+":"+id+":name", r)
	canVote := internal.GetMessage(namePrefix+":"+id+":role", r) != "watch" && name != ""
	templ.Handler(pages.WorkItemSlot(item, parts.Players(id), parts.Watchers(id), votes.Voters(id), votes.All(id), "", true, canVote, baseURL, cards)).ServeHTTP(w, r)
}

func endRound(w http.ResponseWriter, r *http.Request, id, namePrefix string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string) {
	item := store.Get(id)
	wasRevealed := !item.CountdownUntil.IsZero() && time.Since(item.CountdownUntil) > 1500*time.Millisecond
	item.Active = false
	item.CountdownUntil = time.Time{}
	if wasRevealed {
		item.Summary = ""
	}
	store.Set(item)
	votes.Clear(id)
	internal.SessionEvents.Notify(id)
	name := internal.GetMessage(namePrefix+":"+id+":name", r)
	canVote := internal.GetMessage(namePrefix+":"+id+":role", r) != "watch" && name != ""
	templ.Handler(pages.WorkItemSlot(item, parts.Players(id), parts.Watchers(id), votes.Voters(id), votes.All(id), "", true, canVote, baseURL, cards)).ServeHTTP(w, r)
}

func startCountdown(w http.ResponseWriter, r *http.Request, id string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string, namePrefix string) {
	item := store.Get(id)
	if !item.Active {
		http.Error(w, "round not active", http.StatusConflict)
		return
	}
	item.CountdownUntil = time.Now().Add(3 * time.Second)
	store.Set(item)
	internal.SessionEvents.Notify(id)
	name := internal.GetMessage(namePrefix+":"+id+":name", r)
	canVote := internal.GetMessage(namePrefix+":"+id+":role", r) != "watch" && name != ""
	templ.Handler(pages.WorkItemSlot(item, parts.Players(id), parts.Watchers(id), votes.Voters(id), votes.All(id), votes.Get(id, name), true, canVote, baseURL, cards)).ServeHTTP(w, r)
}

func cancelCountdown(w http.ResponseWriter, r *http.Request, id string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string, namePrefix string) {
	item := store.Get(id)
	item.CountdownUntil = time.Time{}
	store.Set(item)
	internal.SessionEvents.Notify(id)
	name := internal.GetMessage(namePrefix+":"+id+":name", r)
	canVote := internal.GetMessage(namePrefix+":"+id+":role", r) != "watch" && name != ""
	templ.Handler(pages.WorkItemSlot(item, parts.Players(id), parts.Watchers(id), votes.Voters(id), votes.All(id), votes.Get(id, name), true, canVote, baseURL, cards)).ServeHTTP(w, r)
}

func castVote(w http.ResponseWriter, r *http.Request, id, namePrefix string, store *internal.WorkItemStore, parts *internal.ParticipantStore, votes *internal.VoteStore, baseURL string, cards []string) {
	name := internal.GetMessage(namePrefix+":"+id+":name", r)
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if internal.GetMessage(namePrefix+":"+id+":role", r) == "watch" {
		http.Error(w, "watchers cannot vote", http.StatusForbidden)
		return
	}
	item := store.Get(id)
	if !item.Active {
		http.Error(w, "round not active", http.StatusConflict)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	value := strings.TrimSpace(r.FormValue("value"))
	if value == "" {
		http.Error(w, "value required", http.StatusBadRequest)
		return
	}
	allowed := false
	for _, c := range cards {
		if c == value {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "invalid card", http.StatusBadRequest)
		return
	}
	votes.Cast(id, name, value)
	templ.Handler(pages.CardPicker(cards, value, baseURL+"/vote")).ServeHTTP(w, r)
}

func themeOrDefault(r *http.Request) string {
	theme := internal.GetMessage("theme", r)
	if theme == "" {
		return defaultTheme
	}
	return theme
}

func absoluteURL(r *http.Request, path string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + path
}

func setupStaticHandlers(router *chi.Mux) {
	router.Get("/styles.css", serveCSS)
	router.Get("/grug.webp", serveGrug)
	router.Get("/office.jpg", serveOffice)
}

func serveCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, "./web/static/css/output.css")
}

func serveGrug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, "./web/static/grug.webp")
}

func serveOffice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, "./web/static/office.jpg")
}

func setup() {
	if dsn := os.Getenv("TURSO_DATABASE"); dsn != "" {
		if err := internal.InitDB(dsn); err != nil {
			log.Println("db init failed:", err)
		}
	}
	if err := internal.SetupSessionManager(); err != nil {
		log.Println("session setup failed:", err)
	}
}
