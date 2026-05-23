package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/robgtest/sprint-tools/internal"
	"github.com/robgtest/sprint-tools/web/pages"
)

const (
	defaultTheme   = "retro"
	synthwaveTheme = "synthwave"
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

	router.Get("/story-pointing", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.StoryPointingPage(themeOrDefault(r))).ServeHTTP(w, r)
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
		isHost := internal.GetMessage("sp:"+id+":host", r) == "1"
		participants := internal.StoryPointingParticipants.List(id)
		workItem := internal.StoryPointingWorkItems.Get(id)
		templ.Handler(pages.StoryPointingSessionPage(theme, sessionURL, id, name, isHost, participants, workItem)).ServeHTTP(w, r)
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
		internal.PutMessage("sp:"+id+":name", name, r)
		internal.StoryPointingParticipants.Add(id, name)
		http.Redirect(w, r, "/story-pointing/"+id, http.StatusSeeOther)
	})

	router.Get("/story-pointing/{id}/participants", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		participants := internal.StoryPointingParticipants.List(id)
		templ.Handler(pages.ParticipantsList(participants)).ServeHTTP(w, r)
	})

	router.Get("/ok-no", func(w http.ResponseWriter, r *http.Request) {
		theme := themeOrDefault(r)
		templ.Handler(pages.OkNoHelpPage(theme)).ServeHTTP(w, r)
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
		isHost := internal.GetMessage("on:"+id+":host", r) == "1"
		participants := internal.OkNoHelpParticipants.List(id)
		workItem := internal.OkNoHelpWorkItems.Get(id)
		templ.Handler(pages.OkNoHelpSessionPage(theme, sessionURL, id, name, isHost, participants, workItem)).ServeHTTP(w, r)
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
		internal.PutMessage("on:"+id+":name", name, r)
		internal.OkNoHelpParticipants.Add(id, name)
		http.Redirect(w, r, "/ok-no/"+id, http.StatusSeeOther)
	})

	router.Get("/ok-no/{id}/participants", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		participants := internal.OkNoHelpParticipants.List(id)
		templ.Handler(pages.ParticipantsList(participants)).ServeHTTP(w, r)
	})

	router.Get("/story-pointing/{id}/work-item", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetStoryPointingSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		item := internal.StoryPointingWorkItems.Get(id)
		templ.Handler(pages.WorkItemCard(item)).ServeHTTP(w, r)
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
		saveWorkItem(w, r, id, internal.StoryPointingWorkItems, "/story-pointing/"+id+"/work-item")
	})

	router.Get("/ok-no/{id}/work-item", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, ok := internal.GetOkNoHelpSession(id); !ok {
			http.NotFound(w, r)
			return
		}
		item := internal.OkNoHelpWorkItems.Get(id)
		templ.Handler(pages.WorkItemCard(item)).ServeHTTP(w, r)
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
		saveWorkItem(w, r, id, internal.OkNoHelpWorkItems, "/ok-no/"+id+"/work-item")
	})

	router.Put("/theme", func(w http.ResponseWriter, r *http.Request) {
		if internal.GetMessage("theme", r) != synthwaveTheme {
			internal.PutMessage("theme", synthwaveTheme, r)
		} else {
			internal.PutMessage("theme", defaultTheme, r)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func saveWorkItem(w http.ResponseWriter, r *http.Request, id string, store *internal.WorkItemStore, saveURL string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	item := &internal.WorkItem{
		SessionID: id,
		Summary:   strings.TrimSpace(r.FormValue("summary")),
	}
	store.Set(item)
	templ.Handler(pages.WorkItemForm(item, saveURL)).ServeHTTP(w, r)
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
}

func serveCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, "./web/static/css/output.css")
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
