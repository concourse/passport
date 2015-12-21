package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/concourse/atc/auth/provider"
	"github.com/pivotal-golang/lager"
)

const OAuthStateCookie = "_concourse_oauth_state"

type OAuthState struct {
	Redirect string `json:"redirect"`
}

type OAuthBeginHandler struct {
	logger     lager.Logger
	providers  provider.Providers
	privateKey *rsa.PrivateKey
}

func NewOAuthBeginHandler(
	logger lager.Logger,
	providers provider.Providers,
	privateKey *rsa.PrivateKey,
) http.Handler {
	return &OAuthBeginHandler{
		logger:     logger,
		providers:  providers,
		privateKey: privateKey,
	}
}

func (handler *OAuthBeginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	providerName := r.FormValue(":provider")

	provider, found := handler.providers[providerName]
	if !found {
		handler.logger.Info("unknown-provider", lager.Data{
			"provider": providerName,
		})

		w.WriteHeader(http.StatusNotFound)
		return
	}

	oauthState, err := json.Marshal(OAuthState{
		Redirect: r.FormValue("redirect"),
	})
	if err != nil {
		handler.logger.Error("failed-to-marshal-state", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encodedState := base64.RawURLEncoding.EncodeToString(oauthState)

	authCodeURL := provider.AuthCodeURL(encodedState)

	http.SetCookie(w, &http.Cookie{
		Name:    OAuthStateCookie,
		Value:   encodedState,
		Path:    "/",
		Expires: time.Now().Add(CookieAge),
	})

	http.Redirect(w, r, authCodeURL, http.StatusTemporaryRedirect)
}
