package connections

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const buffer = 100

func New() *Conns {
	return &Conns{m: map[chan<- *websocket.PreparedMessage]bool{}}
}

type Conns struct {
	m map[chan<- *websocket.PreparedMessage]bool
	sync.Mutex
}

func (c *Conns) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] New connection", r.RemoteAddr)
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		log.Printf("[%s] Failed creating websocket: %s", r.RemoteAddr, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	defer log.Printf("[%s] Disconnected", r.RemoteAddr)

	ch := make(chan *websocket.PreparedMessage, buffer)
	c.add(ch)
	defer c.remove(ch)

	for {
		select {
		case v := <-ch:
			err := conn.WritePreparedMessage(v)
			if err != nil {
				log.Printf("[%s] Failed writing to connection: %s", r.RemoteAddr, err)
				return
			}
		case <-r.Context().Done(): // Wait for client to close the connection.
			log.Printf("[%s] Client closed connection", r.RemoteAddr)
			return
		}
	}
}

func (c *Conns) Write(data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed marshaling %v: %s", data, err)
		return
	}

	p, err := websocket.NewPreparedMessage(1, b)

	fmt.Printf("Sending: %s\n", string(b))

	c.Lock()
	defer c.Unlock()
	for ch := range c.m {
		ch <- p
	}
}

func (c *Conns) add(ch chan *websocket.PreparedMessage) {
	c.Lock()
	defer c.Unlock()
	c.m[ch] = true
}

func (c *Conns) remove(ch chan *websocket.PreparedMessage) {
	c.Lock()
	defer c.Unlock()
	delete(c.m, ch)
}
