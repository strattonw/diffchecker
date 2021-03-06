package diffchecker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	diffcheckerurl        = "https://diffchecker-api-production.herokuapp.com"
	diffcheckersessions   = diffcheckerurl + "/sessions"
	diffcheckerdiffs      = diffcheckerurl + "/diffs"
	diffcheckersuccessurl = "https://www.diffchecker.com/"
	authtokenkey          = "authToken"
)

type DiffChecker struct {
	Email    string
	Password string
}

func (checker DiffChecker) Upload(left string, right string, title string) (diffcheckerurl string, err error) {
	return checker.UploadWithDuration(left, right, title, FOREVER)
}

func (checker DiffChecker) UploadBytes(left []byte, right []byte, title string) (diffcheckerurl string, err error) {
	return checker.UploadBytesWithDuration(left, right, title, FOREVER)
}

func (checker DiffChecker) UploadBytesWithDuration(left []byte, right []byte, title string, expiry DiffCheckerExpiry) (diffcheckerurl string, err error) {
	return checker.UploadWithDuration(string(left), string(right), title, expiry)
}

func (checker DiffChecker) UploadWithDuration(left string, right string, title string, expiry DiffCheckerExpiry) (diffcheckerurl string, err error) {
	token, err := checker.auth()

	if err != nil {
		return "", err
	}

	urlValues := url.Values{
		"left":   {left},
		"right":  {right},
		"expiry": {expiry.String()},
	}

	if len(title) > 0 {
		urlValues.Add("title", title)
	}

	request, _ := http.NewRequest("POST", diffcheckerdiffs, strings.NewReader(urlValues.Encode()))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	response, _ := client.Do(request)
	defer response.Body.Close()

	if response.StatusCode != 201 {
		return "", fmt.Errorf("response was %d, not 201 as expected", response.StatusCode)
	}

	jsonBody, err := parseJson(response.Body)

	if err != nil {
		return "", err
	}

	return diffcheckersuccessurl + jsonBody["slug"].(string), nil
}

func (checker DiffChecker) auth() (token string, err error) {
	response, err := http.PostForm(diffcheckersessions, url.Values{"email": {checker.Email}, "password": {checker.Password}})

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("response was %d, not 200 as expected", response.StatusCode)
	}

	jsonBody, err := parseJson(response.Body)

	if err != nil {
		return "", err
	}

	if jsonBody[authtokenkey] == nil {
		return "", fmt.Errorf("response did not contain %s", authtokenkey)
	}

	return jsonBody[authtokenkey].(string), nil
}

func parseJson(body io.ReadCloser) (jsonBody map[string]interface{}, err error) {
	bytes, err := ioutil.ReadAll(body)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(bytes, &jsonBody)

	return
}
