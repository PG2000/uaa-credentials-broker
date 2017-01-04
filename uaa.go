package main

import (
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/lager"
)

type Users struct {
	Resources    []User
	TotalResults int
}

type User struct {
	ID       string  `json:"id,omitempty"`
	UserName string  `json:"userName,omitempty"`
	Password string  `json:"password,omitempty"`
	Active   bool    `json:"active,omitempty"`
	Emails   []Email `json:"emails"`
}

type Clients struct {
	Resources    []Client
	TotalResults int
}

type Client struct {
	ID                   string   `json:"client_id,omitempty"`
	Name                 string   `json:"name,omitempty"`
	AuthorizedGrantTypes []string `json:"authorized_grant_types,omitempty"`
	Scope                []string `json:"scope,omitempty"`
	RedirectURI          []string `json:"redirect_uri,omitempty"`
	Active               bool     `json:"active,omitempty"`
	ClientSecret         string   `json:"client_secret,omitempty"`
}

type Email struct {
	Value   string `json:"value,omitempty"`
	Primary bool   `json:"primary"`
}

type AuthClient interface {
	GetClient(clientID string) (Client, error)
	CreateClient(client Client) (Client, error)
	DeleteClient(clientID string) error
	GetUser(userID string) (User, error)
	CreateUser(user User) (User, error)
	DeleteUser(userID string) error
}

type UAAClient struct {
	logger   lager.Logger
	client   *http.Client
	endpoint string
	zone     string
}

func (c *UAAClient) GetClient(clientID string) (Client, error) {
	c.logger.Info("uaa-get-client", lager.Data{"clientID": clientID})

	u, _ := url.Parse(fmt.Sprintf("%s/oauth/clients", c.endpoint))
	q := u.Query()
	q.Add("filter", fmt.Sprintf(`client_id eq "%s"`, clientID))
	q.Add("count", "1")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("X-Identity-Zone-Id", c.zone)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return Client{}, err
	}

	clients := Clients{}
	err = decodeBody(resp, &clients)
	if err != nil {
		return Client{}, err
	}

	if clients.TotalResults != 1 {
		return Client{}, fmt.Errorf("Expected to find exactly one client; got %d", clients.TotalResults)
	}

	return clients.Resources[0], nil
}

func (c *UAAClient) CreateClient(client Client) (Client, error) {
	c.logger.Info("uaa-create-client", lager.Data{"clientID": client.Name})

	body, _ := encodeBody(client)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/oauth/clients", c.endpoint), body)
	req.Header.Add("X-Identity-Zone-Id", c.zone)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return Client{}, err
	}

	fmt.Println(resp, resp.Request, client)

	if resp.StatusCode != 201 {
		return Client{}, fmt.Errorf("Expected status 201; got: %d", resp.StatusCode)
	}

	err = decodeBody(resp, &client)
	if err != nil {
		return Client{}, err
	}

	return client, nil
}

func (c *UAAClient) DeleteClient(clientID string) error {
	c.logger.Info("uaa-delete-client", lager.Data{"clientID": clientID})

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/oauth/clients/%s", c.endpoint, clientID), nil)
	req.Header.Add("X-Identity-Zone-Id", c.zone)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected status 200; got: %d", resp.StatusCode)
	}

	return nil
}

func (c *UAAClient) GetUser(userID string) (User, error) {
	c.logger.Info("uaa-get-user", lager.Data{"userID": userID})

	u, _ := url.Parse(fmt.Sprintf("%s/Users", c.endpoint))
	q := u.Query()
	q.Add("filter", fmt.Sprintf(`userName eq "%s"`, userID))
	q.Add("count", "1")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("X-Identity-Zone-Id", c.zone)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return User{}, err
	}

	users := Users{}
	err = decodeBody(resp, &users)
	if err != nil {
		return User{}, err
	}

	if users.TotalResults != 1 {
		return User{}, fmt.Errorf("Expected to find exactly one user; got %d", users.TotalResults)
	}

	return users.Resources[0], nil
}

func (c *UAAClient) CreateUser(user User) (User, error) {
	c.logger.Info("uaa-create-user", lager.Data{"userID": user.UserName})

	body, _ := encodeBody(user)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/Users", c.endpoint), body)
	req.Header.Add("X-Identity-Zone-Id", c.zone)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return User{}, err
	}

	fmt.Println(resp, resp.Request, user)

	if resp.StatusCode != 201 {
		return User{}, fmt.Errorf("Expected status 201; got: %d", resp.StatusCode)
	}

	err = decodeBody(resp, &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (c *UAAClient) DeleteUser(userID string) error {
	c.logger.Info("uaa-delete-user", lager.Data{"userID": userID})

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/Users/%s", c.endpoint, userID), nil)
	req.Header.Add("X-Identity-Zone-Id", c.zone)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected status 200; got: %d", resp.StatusCode)
	}

	return nil
}
