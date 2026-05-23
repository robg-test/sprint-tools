package internal

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/goredisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/redis/go-redis/v9"
)

var UserSessionManager *scs.SessionManager

func getRedisURI() string {
	redisURI := os.Getenv("REDIS_HOST")
	if redisURI == "" {
		redisURI = "localhost"
	}
	return redisURI
}

func SetupSessionManager() error {
	log.Println("Setting up session manager on host: " + getRedisURI())

	opt, err := redis.ParseURL("redis://" + getRedisURI() + ":6379")
	if err != nil {
		return err
	}
	client := redis.NewClient(opt)

	UserSessionManager = scs.New()
	UserSessionManager.Lifetime = 24 * time.Hour
	UserSessionManager.Cookie.Secure = true
	UserSessionManager.Cookie.SameSite = http.SameSiteStrictMode
	UserSessionManager.Store = goredisstore.New(client)

	return nil
}

func PutMessage(key string, value string, r *http.Request) {
	UserSessionManager.Put(r.Context(), key, value)
}

func GetMessage(key string, r *http.Request) string {
	if UserSessionManager == nil {
		return ""
	}
	return UserSessionManager.GetString(r.Context(), key)
}

func ClearSession(r *http.Request) {
	UserSessionManager.Destroy(r.Context())
}
