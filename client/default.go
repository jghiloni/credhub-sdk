package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/jghiloni/credhub-sdk/auth"
)

type client struct {
	url string
	hc  *http.Client
}

// New creates a new Credhub client with OAuth2 authentication
func New(credhubURL, clientID, clientSecret string, skipTLSVerify bool) (Credhub, error) {
	var cli *http.Client
	var err error
	if cli, err = auth.NewOAuthClient(credhubURL, clientID, clientSecret, skipTLSVerify); err != nil {
		return nil, err
	}

	return &client{
		url: credhubURL,
		hc:  cli,
	}, nil
}

func (c *client) FindByPath(path string) ([]Credential, error) {
	var retBody struct {
		Credentials []Credential `json:"credentials"`
	}

	resp, err := c.hc.Get(c.url + "/api/v1/data?path=" + path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, errors.New("Path Not Found")
	}

	marshaller := json.NewDecoder(resp.Body)

	err = marshaller.Decode(&retBody)
	return retBody.Credentials, err
}

func (c *client) GetAllByName(name string) ([]Credential, error) {
	return c.getByName(name, false, -1)
}

func (c *client) GetVersionsByName(name string, numVersions int) ([]Credential, error) {
	return c.getByName(name, false, numVersions)
}

func (c *client) GetLatestByName(name string) (Credential, error) {
	creds, err := c.getByName(name, true, -1)
	if err != nil {
		return Credential{}, err
	}

	return creds[0], nil
}

func (c *client) getByName(name string, latest bool, numVersions int) ([]Credential, error) {
	var retBody struct {
		Data []Credential `json:"data"`
	}

	chURL := c.url + "/api/v1/data?"

	params := url.Values{}
	params.Add("name", name)

	if latest {
		params.Add("current", "true")
	}

	if numVersions > 0 {
		params.Add("versions", fmt.Sprint(numVersions))
	}

	chURL += params.Encode()
	resp, err := c.hc.Get(chURL)
	if err != nil {
		return retBody.Data, err
	}

	if resp.StatusCode == 404 {
		return nil, errors.New("Name Not Found")
	}

	marshaller := json.NewDecoder(resp.Body)

	err = marshaller.Decode(&retBody)
	if err != nil {
		return nil, err
	}

	data := retBody.Data
	sort.Slice(data, func(i, j int) bool {
		less := strings.Compare(data[i].Created, data[j].Created)
		// we want to sort in reverse order, so return the opposite of what you'd normally do
		return less > 0
	})

	return retBody.Data, err
}

func (c *client) Set(credential Credential) (Credential, error) {
	return Credential{}, nil
}
