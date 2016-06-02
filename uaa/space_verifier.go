package uaa

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/oauth2"

	"github.com/dgrijalva/jwt-go"
	"github.com/pivotal-golang/lager"
)

type SpaceVerifier struct {
	spaceGUIDs []string
	cfAPIURL   string
}

func NewSpaceVerifier(
	spaceGUIDs []string,
	cfAPIURL string,
) SpaceVerifier {
	return SpaceVerifier{
		spaceGUIDs: spaceGUIDs,
		cfAPIURL:   cfAPIURL,
	}
}

type UAAToken struct {
	UserID string `json:"user-id"`
}

type CFSpaceDevelopersResponse struct {
	UserInfos []CFUserInfo `json:"resources"`
}

type CFUserInfo struct {
	Metadata struct {
		GUID string `json:"guid"`
	} `json:"metadata"`
}

func (verifier SpaceVerifier) Verify(logger lager.Logger, httpClient *http.Client) (bool, error) {
	oauth2Transport, ok := httpClient.Transport.(*oauth2.Transport)
	if !ok {
		return false, errors.New("httpClient transport must be of type oauth2.Transport")
	}

	token, err := oauth2Transport.Source.Token()
	if err != nil {
		return false, err
	}

	tokenParts := strings.Split(token.AccessToken, ".")
	if len(tokenParts) < 2 {
		return false, errors.New("access token contains an invalid number of segments")
	}

	decodedClaims, err := jwt.DecodeSegment(tokenParts[1])
	if err != nil {
		return false, err
	}

	var uaaToken UAAToken
	err = json.Unmarshal(decodedClaims, &uaaToken)
	if err != nil {
		return false, err
	}

	cfAPIURL, err := url.Parse(verifier.cfAPIURL)
	if err != nil {
		return false, err
	}

	for _, verifierSpaceGUID := range verifier.spaceGUIDs {
		cfAPIURL.Path = path.Join("v2", "spaces", verifierSpaceGUID, "developers")
		response, err := httpClient.Get(cfAPIURL.String())
		if err != nil {
			return false, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return false, fmt.Errorf("unexpected response code from CF API URL: %d", response.StatusCode)
		}

		var cfSpaceDevelopersResponse CFSpaceDevelopersResponse
		err = json.NewDecoder(response.Body).Decode(&cfSpaceDevelopersResponse)
		if err != nil {
			return false, err
		}

		for _, userInfo := range cfSpaceDevelopersResponse.UserInfos {
			if userInfo.Metadata.GUID == uaaToken.UserID {
				return true, nil
			}
		}
	}

	logger.Info("not-in-spaces", lager.Data{
		"want": verifier.spaceGUIDs,
	})

	return false, nil
}
