package auth

import (
	"net/http"

	"github.com/gorilla/securecookie"
)

const cookieName = "kiln-session"

// Session holds the data stored in the signed cookie.
type Session struct {
	Login  string `json:"login"`
	Avatar string `json:"avatar"`
}

// SessionStore manages encoding and decoding sessions into signed cookies.
type SessionStore struct {
	sc *securecookie.SecureCookie
}

// NewSessionStore creates a SessionStore using the provided secret as the hash key.
func NewSessionStore(secret string) *SessionStore {
	return &SessionStore{
		sc: securecookie.New([]byte(secret), nil),
	}
}

// Save encodes the session and sets it as a signed cookie on the response.
func (s *SessionStore) Save(w http.ResponseWriter, session *Session) error {
	values := map[string]string{
		"login":  session.Login,
		"avatar": session.Avatar,
	}
	encoded, err := s.sc.Encode(cookieName, values)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
	return nil
}

// Get reads the session cookie from the request and decodes it.
func (s *SessionStore) Get(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	if err := s.sc.Decode(cookieName, cookie.Value, &values); err != nil {
		return nil, err
	}
	return &Session{
		Login:  values["login"],
		Avatar: values["avatar"],
	}, nil
}

// Clear expires the session cookie.
func (s *SessionStore) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
