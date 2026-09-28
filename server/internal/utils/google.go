package utils

import (
	"fmt"
	"io"
	"net/http"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func GetUserDataFromGoogle(accessToken string) ([]byte, error) {

	response, err := http.Get(oauthGoogleUrlAPI + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}

	return contents, nil
}
