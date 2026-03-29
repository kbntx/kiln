package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"

	ghclient "github.com/kbntx/kiln/internal/github"
)

const oauthStateCookie = "oauth-state"

// OAuthHandler handles GitHub OAuth login, callback, and logout.
type OAuthHandler struct {
	config       *oauth2.Config
	sessions     *SessionStore
	allowedOrg   string
	githubClient ghclient.GitHubClient
	devMode      bool
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(clientID, clientSecret, baseURL, allowedOrg string, sessions *SessionStore, ghClient ghclient.GitHubClient, devMode bool) *OAuthHandler {
	return &OAuthHandler{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     oauthgithub.Endpoint,
			RedirectURL:  baseURL + "/auth/callback",
			Scopes:       []string{"read:org"},
		},
		sessions:     sessions,
		allowedOrg:   allowedOrg,
		githubClient: ghClient,
		devMode:      devMode,
	}
}

// HandleLogin redirects the user to GitHub for authentication.
// In dev mode it creates a session immediately and redirects to "/".
func (h *OAuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if h.devMode {
		_ = h.sessions.Save(w, &Session{
			Login:  "dev-user",
			Avatar: "",
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	state, err := randomState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
	})

	url := h.config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// githubUser is the subset of fields we read from the GitHub user API.
type githubUser struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

// HandleCallback handles the OAuth callback from GitHub.
func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state.
	stateCookie, err := r.Cookie(oauthStateCookie)
	if err != nil || stateCookie.Value == "" {
		http.Error(w, "missing oauth state cookie", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}
	// Clear the state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:   oauthStateCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange code for token.
	code := r.URL.Query().Get("code")
	token, err := h.config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "oauth exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch GitHub user info.
	user, err := fetchGitHubUser(token.AccessToken)
	if err != nil {
		http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check organisation membership using the user's own OAuth token.
	if h.allowedOrg != "" {
		orgs, err := fetchUserOrgs(token.AccessToken)
		if err != nil {
			http.Error(w, "failed to fetch user orgs: "+err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("org membership check", "user", user.Login, "userOrgs", orgs, "allowedOrg", h.allowedOrg)
		if !slices.Contains(orgs, h.allowedOrg) {
			http.Error(w, "user is not a member of the required organisation", http.StatusForbidden)
			return
		}
	}

	// Create session and redirect.
	if err := h.sessions.Save(w, &Session{
		Login:  user.Login,
		Avatar: user.AvatarURL,
	}); err != nil {
		http.Error(w, "failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// fetchGitHubUser calls the GitHub API to get the authenticated user's profile.
func fetchGitHubUser(accessToken string) (*githubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// fetchUserOrgs calls the GitHub API to list the authenticated user's organisations.
func fetchUserOrgs(accessToken string) ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/orgs", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var orgs []struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, err
	}

	names := make([]string, len(orgs))
	for i, o := range orgs {
		names[i] = o.Login
	}
	return names, nil
}

// randomState generates a cryptographically random hex string for the OAuth state parameter.
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
