package provider

import (
	"net/http"

	"github.com/pivotal-golang/lager"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type Providers map[string]Provider

//go:generate counterfeiter . Provider

type Provider interface {
	DisplayName() string

	OAuthClient
	Verifier
}

type OAuthClient interface {
	AuthCodeURL(string, ...oauth2.AuthCodeOption) string
	Exchange(context.Context, string) (*oauth2.Token, error)
	Client(context.Context, *oauth2.Token) *http.Client
}

//go:generate counterfeiter . Verifier

type Verifier interface {
	Verify(lager.Logger, *http.Client) (bool, error)
}
