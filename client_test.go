package loggly_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/treetopllc/go-loggly"
)

type Message struct {
	id int
}

func (m Message) String() string {
	return fmt.Sprintf(`
{
	"id": "%v",
	"log": "Message
strings: %v",
	"arry": ["your", "a", "wizzard!"]
}`, m.id, strings.Repeat("TreeTop", m.id%4))

}

func TestClient(t *testing.T) {
	c := loggly.NewClient("<your token here>", "noblehack", "go3-big")
	count := 10
	sleep := time.Millisecond * 200
	start := time.Now()
	for i := 0; i < count; i++ {
		c.Send([]byte(Message{i}.String()))
		time.Sleep(sleep)
	}
	end := time.Now()
	fmt.Printf("Took %v per\n", (end.Sub(start))/time.Duration(count)-sleep)
}
