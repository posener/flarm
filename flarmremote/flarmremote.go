package flarmremote

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/posener/flarm/process"
)

// Open connects to a remote falrm server, and returns a an object that implements flarmReader..
func Open(addr string) (*Conn, error) {
	d := websocket.Dialer{}
	conn, _, err := d.Dial(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed dialing %s: %v", addr, err)
	}

	return &Conn{conn: conn}, nil
}

type Conn struct {
	conn *websocket.Conn
}

func (c *Conn) Range(f func(interface{})) error {
	defer c.conn.Close()
	for {
		v, err := c.next()
		if err != nil {
			return err
		}
		f(v)
	}
}

// next is used in Range and exists for testing purposes.
func (c *Conn) next() (interface{}, error) {
	var o process.Object
	err := c.conn.ReadJSON(&o)
	return &o, err
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
