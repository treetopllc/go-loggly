package loggly

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/treetopllc/go-loggly/async"
)

//Does not parse json properly when using bulk endpoint
const api = "https://logs-01.loggly.com/inputs/"
const Version = "0.0.1"

type Entry interface {
	String() string
}

type Client interface {
	Write(p []byte) (n int, err error)
	Flush()
	Send(...Entry)
	Close()
}

type client struct {
	token string
	tags  []string
	queue async.Queue
}

func NewClient(token string, tags ...string) Client {
	c := &client{
		token: token,
		tags:  tags,
	}
	c.queue = async.NewQueue(c.writemsgs, 100, time.Millisecond)
	return c
}

func (c *client) Write(p []byte) (int, error) {
	c.queue.Add(p)
	return len(p), nil
}

//loosely based off of https://github.com/segmentio/go-loggly
func (c *client) writemsgs(msgs [][]byte) {
	var wg sync.WaitGroup
	for _, msg := range msgs {
		body := msg
		wg.Add(1)
		go func() {
			req, err := http.NewRequest("POST", api+c.token, bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("Error writing to loggly: %v", err)
			}

			req.Header.Add("User-Agent", "tt: go-loggly ("+Version+")")
			req.Header.Add("Content-Type", "text/plain")
			req.Header.Add("Content-Length", string(len(body)))

			if len(c.tags) != 0 {
				req.Header.Add("X-Loggly-Tag", strings.Join(c.tags, ","))
			}

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				fmt.Println("Error writing to loggly: %v", err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				resp, _ := ioutil.ReadAll(res.Body)
				fmt.Println("Error writing to loggly: %v (%v)", resp, res.StatusCode)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (c *client) Flush() {
	c.queue.Flush()
}

func (c *client) Send(entries ...Entry) {
	for _, entry := range entries {
		c.queue.Add([]byte(entry.String()))
	}
}

func (c *client) Close() {
	c.queue.Close()
}
