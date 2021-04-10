package flarmport

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// Remote connects to a remote flarm server, and returns a an object that implements flarmReader..
func Remote(addr string) (*Conn, error) {
	d := websocket.Dialer{
		HandshakeTimeout: time.Second * 10,
	}
	conn, _, err := d.Dial(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed dialing %s: %v", addr, err)
	}

	return &Conn{conn: conn}, nil
}

type Conn struct {
	conn *websocket.Conn
}

func (c *Conn) Range(ctx context.Context, f func(Data)) error {
	defer c.conn.Close()
	for ctx.Err() == nil {
		v, err := c.next()
		if err != nil {
			return err
		}
		if ctx.Err() == nil {
			f(v)
		}
	}
	return ctx.Err()
}

// next is used in Range and exists for testing purposes.
func (c *Conn) next() (Data, error) {
	var o Data
	err := c.conn.ReadJSON(&o)
	return o, err
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
