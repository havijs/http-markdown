package app

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type key int

const (
	sessionTokenKey key = iota
)

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := getSessionCookie(r)
		if c == nil {
			sessionToken := uuid.NewString()
			expiry := time.Now().Add(24 * time.Hour)
			app.sessions[sessionToken] = &Session{
				expiry:   expiry,
				loggedIn: false,
				prevPage: "/",
			}
			c = &http.Cookie{
				Name:    "session_token",
				Value:   sessionToken,
				Expires: expiry,
			}
			http.SetCookie(w, c)
		}
		ctx := context.WithValue(r.Context(), sessionTokenKey, c.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getSession(r *http.Request) *Session {
	sessionToken := r.Context().Value(sessionTokenKey).(string)
	session := app.sessions[sessionToken]
	return session
}

func getSessionCookie(r *http.Request) *http.Cookie {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return nil
		}
		// For any other type of error, return a bad request status
		return nil
	}
	sessionToken := c.Value
	userSession, exists := app.sessions[sessionToken]
	if !exists {
		// If the session token is not present in session map, return an unauthorized error
		return nil
	}
	// If the session is present, but has expired, we can delete the session, and return
	// an unauthorized status
	if userSession.isExpired() {
		delete(app.sessions, sessionToken)
		return nil
	}
	return c
}

func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if !session.loggedIn {
			session.prevPage = r.URL.Path
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
