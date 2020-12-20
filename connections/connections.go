package connections

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

func New() *Conns {
	return &Conns{m: map[*websocket.Conn]bool{}}
}

type Conns struct {
	m map[*websocket.Conn]bool
	sync.Mutex
}

func (c *Conns) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		log.Printf("Failed creating connection from %s: %s", r.RemoteAddr, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer c.remove(conn)

	c.add(conn)

	<-r.Context().Done() // Wait for client to close the connection.
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

	// Write the message to all current connections. Remember all fails writes in order to delete
	// them later on.
	fails := map[*websocket.Conn]error{}
	for w := range c.m {
		w := w
		err := w.WritePreparedMessage(p)
		if err != nil {
			fails[w] = err
		}
	}

	// Delete all stale connections.
	for w, err := range fails {
		log.Printf("Removing failed connection: %s: %s", w.RemoteAddr(), err)
		delete(c.m, w)
	}
}

func (c *Conns) add(w *websocket.Conn) {
	c.Lock()
	defer c.Unlock()
	log.Printf("New connection: %s", w.RemoteAddr())
	c.m[w] = true
}

func (c *Conns) remove(w *websocket.Conn) {
	c.Lock()
	defer c.Unlock()
	log.Printf("Removed connection: %s", w.RemoteAddr())
	delete(c.m, w)
}
