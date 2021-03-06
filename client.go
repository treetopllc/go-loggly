package loggly

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

//Does not parse json properly when using bulk endpoint
const api = "https://logs-01.loggly.com/inputs/"
const Version = "0.0.1"

type Client interface {
	Send([]byte)
}

type client struct {
	token string
	tags  []string
}

func NewClient(token string, tags ...string) Client {
	c := &client{
		token: token,
		tags:  tags,
	}
	return c
}

func (c *client) Send(body []byte) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Panic writing to loggly: %s", err)
		}
	}()
	req, err := http.NewRequest("POST", api+c.token, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error creating loggly request: %v\n", err)
		return
	}

	req.Header.Add("User-Agent", "tt: go-loggly ("+Version+")")
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))

	if len(c.tags) != 0 {
		req.Header.Add("X-Loggly-Tag", strings.Join(c.tags, ","))
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending to loggly: %v\n", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		resp, _ := ioutil.ReadAll(res.Body)
		fmt.Printf("Error response from loggly: %s (%v)\n", string(resp), res.StatusCode)
	}
}
