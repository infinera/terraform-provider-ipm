package ipm_pf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/martian/v3/log"
)

// HostURL - Default Hashicups URL
const HostURL string = "http://localhost:19090"

// Client -
type Client struct {
	HostURL       string
	HTTPClient    *http.Client
	Token         string
	Auth          AuthStruct
	IPMmap     map[string]string
	GetTimeout    time.Duration
	DeleteTimeout time.Duration
	UpdateTimeout time.Duration
}

// AuthStruct -
type AuthStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse -
type AuthResponse struct {
	Token         string `json:"acess_token"`
	Id_token      string `json:"id_token"`
	Refresh_token string `json:"refresh_token"`
	Scope         string `json:"scope"`
	Token_type    string `json:"token_type"`
}

// NewClient -
func NewClient(host, username, password *string) (*Client, error) {
	getTimeout, err := strconv.Atoi(os.Getenv("GET_TIMEOUT"))
	if err != nil {
		getTimeout = 0
	}
	updateTimeout, err := strconv.Atoi(os.Getenv("UPDATE_TIMEOUT"))
	if err != nil {
		updateTimeout = 4
	}
	deleteTimeout, err := strconv.Atoi(os.Getenv("DELETE_TIMEOUT"))
	if err != nil {
		deleteTimeout = 5
	}

	log.Debugf("NewClient: getTimeout = %d, updateTimeout = %d, deleteTimeout = %d", getTimeout, updateTimeout, deleteTimeout)


	c := Client{
		HTTPClient: &http.Client{Timeout: time.Duration(getTimeout) * time.Second},
		// Default Hashicups URL
		HostURL: HostURL,
		Auth: AuthStruct{
			Username: *username,
			Password: *password,
		},
		//UpdateTimeout: time.Duration(updateTimeout) * time.Second,
		//GetTimeout:    time.Duration(getTimeout) * time.Second,
		//DeleteTimeout: time.Duration(deleteTimeout) * time.Second,
	}

	if host != nil {
		c.HostURL = *host
	}

	fmt.Println("NewClient: Signin")

	ar, err := c.SignIn()
	if err != nil {
		return nil, err
	}


	c.Token = ar.Token
	//fmt.Println("ar Token:" + ar.Token)
	
	fmt.Println("c Token:" + c.Token)

	log.SetLevel(log.Debug)

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", c.Token)

	if req.Method == "GET" {
		c.HTTPClient.Timeout = c.GetTimeout
	} else if req.Method != "DELETE" {
		c.HTTPClient.Timeout = c.UpdateTimeout
	} else {
		c.HTTPClient.Timeout = c.DeleteTimeout
	}

	log.Debugf("doRequest: method = %s, Timeout = %v", req.Method, c.HTTPClient.Timeout)

	res, err := c.HTTPClient.Do(req)

	if err != nil {
		log.Debugf("doRequest: Send HTTP Request error %v", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debugf("doRequest: Can not read Reponse Body. error %v", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

// executes commands on IPM
func (c *Client) ExecuteIPMHttpCommand(command, commanduri string, commandBody []byte) (result []byte,  err error) {

	log.Debugf("ExecuteIPMHttpCommand:New HTTP Request https://%s/api/v1%s", c.HostURL, commanduri)
	// fmt.Println("IPMid:", IPMid, "command body"+string(commandBody))
	req, err := http.NewRequest(command, fmt.Sprintf("https://%s/api/v1%s", c.HostURL, commanduri), bytes.NewBuffer(commandBody))
	log.Debugf("ExecuteIPMHttpCommand: Create HTTP Request %v", req)
	if err != nil {
		log.Errorf("ExecuteIPMHttpCommand: IPM ID = %s, Create New HTTP Request failed error %v", err)
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		log.Errorf("ExecuteIPMHttpCommand: Send HTTP NewRequest failed for devide  error= %v", err)
		return nil, err
	}
	log.Debugf("ExecuteIPMHttpCommand: Send HTTP Request SUCCESS. Response = %s", string(body))
	return body, err
}