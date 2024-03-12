package cfg

import (
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
)

// InitializeSessionStore initializes the session store with environment-specific configurations.
func InitializeSessionStore() *sessions.CookieStore {
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("SESSION_KEY is not set")
	}

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	return store
}
