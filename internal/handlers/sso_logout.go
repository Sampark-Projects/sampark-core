package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"windshift/internal/sso"
)

// ssoLogoutTimeout bounds the background call that ends the IdP session.
const ssoLogoutTimeout = 15 * time.Second

// ssoLogoutHint is what the sso_logout cookie carries between login and logout.
type ssoLogoutHint struct {
	Slug         string `json:"p"`
	RefreshToken string `json:"r"`
}

// setLogoutHint remembers which OIDC provider created the session and the
// refresh token that lets logout end the IdP session server-side.
func (h *SSOHandler) setLogoutHint(w http.ResponseWriter, r *http.Request, slug, refreshToken string, rememberMe bool) {
	if refreshToken == "" {
		return
	}
	raw, err := json.Marshal(ssoLogoutHint{Slug: slug, RefreshToken: refreshToken})
	if err != nil {
		return
	}
	if err := h.sessionManager.SetSSOLogoutCookie(w, r, string(raw), rememberMe); err != nil {
		slog.Warn("failed to store SSO logout hint", slog.String("component", "sso"), slog.String("provider", slug), slog.Any("error", err))
	}
}

// EndIdPSession runs on app logout. For sessions created through OIDC SSO it
// ends the IdP session through the back channel (refresh token sent to the
// end-session endpoint) so the browser stays on the app instead of being
// redirected through the IdP and any upstream IdP it brokers to. The next SSO
// login then shows the IdP's login page again.
func (h *SSOHandler) EndIdPSession(w http.ResponseWriter, r *http.Request) {
	raw, err := h.sessionManager.GetSSOLogoutCookie(r)
	if err != nil {
		return
	}
	h.sessionManager.ClearSSOLogoutCookie(w, r)

	var hint ssoLogoutHint
	if err := json.Unmarshal([]byte(raw), &hint); err != nil || hint.Slug == "" || hint.RefreshToken == "" {
		return
	}
	callbackURI := h.getRedirectURI(r, hint.Slug)

	// The response must not wait on the IdP; run the call in the background.
	go h.backchannelLogout(hint, callbackURI)
}

func (h *SSOHandler) backchannelLogout(hint ssoLogoutHint, callbackURI string) {
	ctx, cancel := context.WithTimeout(context.Background(), ssoLogoutTimeout)
	defer cancel()
	logAttrs := []any{slog.String("component", "sso"), slog.String("provider", hint.Slug)}

	provider, err := h.providerStore.GetBySlug(hint.Slug)
	if err != nil || provider == nil || !provider.Enabled || provider.ProviderType != sso.ProviderTypeOIDC {
		return
	}
	clientSecret, err := h.encryption.Decrypt(provider.ClientSecretEncrypted)
	if err != nil {
		return
	}
	relyingParty, err := h.oidcService.CreateRelyingParty(ctx, provider, callbackURI, clientSecret)
	if err != nil {
		slog.Warn("SSO logout: provider discovery failed", append(logAttrs, slog.Any("error", err))...)
		return
	}
	endpoint := relyingParty.GetEndSessionEndpoint()
	if endpoint == "" {
		return
	}

	form := url.Values{
		"client_id":     {provider.ClientID},
		"client_secret": {clientSecret},
		"refresh_token": {hint.RefreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := relyingParty.HttpClient().Do(req)
	if err != nil {
		slog.Warn("SSO logout: end-session request failed", append(logAttrs, slog.Any("error", err))...)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	// A rejected refresh token usually means the IdP session already expired.
	if resp.StatusCode >= 300 {
		slog.Warn("SSO logout: IdP rejected end-session request", append(logAttrs, slog.Int("status", resp.StatusCode))...)
		return
	}
	slog.Info("SSO logout: IdP session ended", logAttrs...)
}
