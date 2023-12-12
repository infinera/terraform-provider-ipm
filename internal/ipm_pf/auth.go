package ipm_pf

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// SignIn - Get a new token for user
func (c *Client) SignIn() (*AuthResponse, error) {

	fmt.Printf("SignIn: host = %s, user= %s, password = %s", c.HostURL, c.Auth.Username, c.Auth.Password)
	if c.Auth.Username == "" || c.Auth.Password == "" {
		return nil, fmt.Errorf("please specify username and password for IPM server")
	}
	/*data := make(map[string]interface{})
	data["username"] = c.Auth.Username
	data["password"] = c.Auth.Password
	data["grant_type"] = "password"
	data["client_secret"] = "xr-web-client"
	data["client_id"] = "xr-web-client"
	fmt.Printf("SignIn: data = %s", data)

	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	fmt.Printf("\nSignIn: rb = %s", rb)*/

	// req, err := http.NewRequest("POST", fmt.Sprintf("%s/signin", c.HostURL), strings.NewReader(string(rb)))
	// fmt.Println("Connecting server: ", fmt.Sprintf("%s/oauth/token?client_id=test&audience=test", c.HostURL))

	// TODO - Need to data drive config with self signed TLS vs valid certificates
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Self signed certificate flag, TODO to remove or data drive it latter

	//url := "https://pt-xrivk824-dv/realms/xr-cm/protocol/openid-connect/token"
	url := "https://" + c.HostURL +"/realms/xr-cm/protocol/openid-connect/token"
  method := "POST"

  //payload := strings.NewReader("username=xr-user-1&password=xr&grant_type=password&client_secret=xr-web-client&client_id=xr-web-client")
	payload := strings.NewReader("username="+ c.Auth.Username+"&password="+c.Auth.Password+"&grant_type=password&client_secret=xr-web-client&client_id=xr-web-client")
	fmt.Printf("payload= %s", payload)

  req, err := http.NewRequest(method, url, payload)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

  res, err := c.HTTPClient.Do(req)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }
  defer res.Body.Close()

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }
  fmt.Printf("body = %s", string(body))

	/*http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Self signed certificate flag, TODO to remove or data drive it latter
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/realms/xr-cm/protocol/openid-connect/token", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("\nSignIn: set header ")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}*/

	//str1 := fmt.Sprintf("%s", body)
	//fmt.Println("body auth" + str1)
	ar := AuthResponse{}
	var respmap = make(map[string]interface{})
	err = json.Unmarshal(body, &respmap)
	if err != nil {
		fmt.Printf("Unmarshal failed")
		return nil, err
	}
	ar.Token = "Bearer " + respmap["access_token"].(string)
	// fmt.Println("token", ar.Token)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

// SignOut - Revoke the token for a user
func (c *Client) SignOut() error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/signout", c.HostURL), strings.NewReader(string("")))
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if string(body) != "Signed out user" {
		return errors.New(string(body))
	}

	return nil
}
