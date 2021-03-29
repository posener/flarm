package flarmremote

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/posener/flarm/process"
	"github.com/posener/wsbeam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	in := wsbeam.New()
	s := httptest.NewServer(in)

	f, err := Open(strings.Replace(s.URL, "http://", "ws://", 1))
	if err != nil {
		t.Fatal(err)
	}

	objs := []*process.Object{{Name: "1"}, {Name: "2"}, {Name: "3"}}

	for _, o := range objs {
		assert.NoError(t, in.Send(o))
	}

	for _, want := range objs {
		got, err := f.next()
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
	f.Close()
	_, err = f.next()
	assert.Error(t, err)
}
