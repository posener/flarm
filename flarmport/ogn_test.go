package flarmport

import (
	"bufio"
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testData = `
Connection closed by foreign host.
@@@ Child "./ogn-decode" started at: Thu Apr 22 13:40:56 2021
@@@ 0 user(s) and 0 logger(s) connected (plus you)
0.802sec:916.200MHz:   1:2:DDFD1D 104436: [ +32.59657, +35.23525]deg    77m  +0.1m/s   0.4m/s 123.1deg  -1.2deg/s 1m1 07x05m Fn:35___ +0.41kHz 45.0/55.5dB/0  0e     0.0km 161.7deg +12.9deg          
0.802sec:916.200MHz: 2:2:123456 104436: [+32.5, +35.1]deg 10m -3.1m/s 0.4m/s 123.1deg -1.2deg/s
`

func TestOGN(t *testing.T) {
	t.Parallel()

	l, err := net.ListenTCP("tcp", nil)
	require.NoError(t, err)

	go func() {
		defer l.Close()
		conn, err := l.AcceptTCP()
		require.NoError(t, err)
		s := bufio.NewScanner(bytes.NewReader([]byte(testData)))
		var i int
		for s.Scan() {
			t.Logf("Line %d: %q", i, s.Text())
			i++
			_, err := conn.Write(append(s.Bytes(), '\n'))
			require.NoError(t, err)
		}
		require.NoError(t, s.Err())
	}()

	ogn, err := OpenOGN(l.Addr().String(), StationInfo{IDMap: map[string]string{"123456": "4X-APL"}})
	require.NoError(t, err)

	for i := 0; i < 4; i++ {
		_, ok := ogn.next()
		assert.False(t, ok)
	}

	// First line of data
	got, ok := ogn.next()
	want := &Data{
		Type:     "glider",
		Name:     "DDFD1D",
		Lat:      32.59657,
		Long:     35.23525,
		Alt:      77,
		Climb:    0.1,
		Dir:      123,
		TurnRate: -1.2,
	}
	assert.True(t, ok)
	got.Time = time.Time{} // Clear time before comparing.
	assert.Equal(t, want, got)

	// Second line of data
	got, ok = ogn.next()
	want = &Data{
		Type:     "towplane",
		Name:     "4X-APL",
		Lat:      32.5,
		Long:     35.1,
		Alt:      10,
		Climb:    -3.1,
		Dir:      123,
		TurnRate: -1.2,
	}
	assert.True(t, ok)
	got.Time = time.Time{} // Clear time before comparing.
	assert.Equal(t, want, got)
}
