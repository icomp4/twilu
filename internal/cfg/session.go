package cfg

import (
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

// InitializeSessionStore initializes the session store with environment-specific configurations.
func InitializeSessionStore() *sessions.CookieStore {
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}

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
