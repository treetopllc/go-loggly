package async

import (
	"sync"
	"time"
)

type Message []byte

type FlushFunc func([][]byte)

type Queue interface {
	Flush()
	Add(...Message)
	Close()
}

type queue struct {
	sync.Mutex
	interval time.Duration
	flusher  FlushFunc
	buffer   [][]byte
	inner    chan Message
	wg       sync.WaitGroup
}

func NewQueue(flusher FlushFunc, capacity uint, interval time.Duration) Queue {
	q := &queue{
		interval: interval,
		flusher:  flusher,
		buffer:   make([][]byte, 0, capacity+1),
		inner:    make(chan Message, capacity),
	}
	go q.open()
	return q
}

func (q *queue) Flush() {
	if len(q.buffer) == 0 {
		return
	}
	q.Lock()
	q.wg.Add(1)
	tmp := make([][]byte, len(q.buffer))
	copy(tmp, q.buffer)
	q.buffer = make([][]byte, 0, cap(q.buffer))

	go func() {
		q.flusher(tmp)
		q.wg.Done()
	}()
	q.Unlock()
}

func (q *queue) Add(msgs ...Message) {
	for _, msg := range msgs {
		q.inner <- msg
	}
}

func (q *queue) open() {
	count := 0
	for {
		timeout := time.After(q.interval)
		select {
		case msg, ok := <-q.inner:
			if !ok {
				q.Flush()
				return
			}
			q.Lock()
			q.buffer = append(q.buffer, msg)
			q.Unlock()
			count++
		case <-timeout:
			q.Flush()
		}
		//Flush when close to full
		if cap(q.buffer)*9/10 <= len(q.buffer) {
			q.Flush()
		}
	}
}

func (q *queue) Close() {
	close(q.inner)
	q.wg.Wait()
}
