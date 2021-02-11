package flarmport

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"testing"

	"github.com/adrianmo/go-nmea"
	"github.com/jacobsa/go-serial/serial"
	"github.com/stretchr/testify/assert"
)

var reSocatPTY = regexp.MustCompile("N PTY is (.+)$")

func TestOpen(t *testing.T) {
	// Use socat to create a pair of ports.
	socat := exec.Command("socat", "-d", "-d", "pty,raw,echo=0", "pty,raw,echo=0")
	fmt.Println("Starting socat...")
	stderr, err := socat.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	err = socat.Start()
	if err != nil {
		t.Fatal(err)
	}

	defer socat.Process.Kill()

	// Get port names from the stderr of socat.
	wPort, rPort := parsePorts(stderr)

	// Create port writer.
	w, err := serial.Open(serial.OpenOptions{
		PortName:        wPort,
		BaudRate:        baudRate,
		MinimumReadSize: 1,
		StopBits:        1,
		DataBits:        8,
		ParityMode:      serial.PARITY_NONE,
	})
	if err != nil {
		t.Fatal(err)
	}

	flarm, err := Open(rPort)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		in   string
		want interface{}
	}{
		{
			name: "PFLAA",
			in:   "$PFLAA,0,-1388,-330,465,2,DD8E8B,78,,44,2.8,2*6F",
			want: TypePFLAA{
				AlarmLevel:       0,
				RelativeNorth:    -1388,
				RelativeEast:     -330,
				RelativeVertical: 465,
				IDType:           "flarm id",
				ID:               "DD8E8B",
				Track:            78,
				TurnRate:         0,
				GroundSpeed:      44,
				ClimbRate:        2.8,
				AircraftType:     "towplane",
			},
		},
		{
			name: "PFLAU",
			in:   "$PFLAU,2,1,1,1,0,,0,,*61",
			want: TypePFLAU{
				Rx:               2,
				Tx:               "no transmission",
				GPS:              "valid on ground",
				Power:            1,
				AlarmLevel:       0,
				RelativeBearing:  0,
				AlarmType:        "no alarm",
				RelativeVertical: 0,
				RelativeDistance: 0,
				ID:               "",
			},
		},
		{
			name: "PGRMZ",
			in:   "$PGRMZ,78,F,2*05",
			want: TypePGRMZ{
				Altitude: 78,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.Write([]byte(tt.in + "\n\r"))
			v, ok := flarm.next()
			assert.True(t, ok)
			assert.Equal(t, tt.want, clean(v))
		})
	}
}

// parsePorts parses socat stderr and returns the opened ports.
func parsePorts(stderr io.Reader) (p1, p2 string) {
	s := bufio.NewScanner(stderr)
	var ports []string
	for len(ports) < 2 {
		if !s.Scan() {
			panic("scanning socat stderr...")
		}
		if groups := reSocatPTY.FindStringSubmatch(s.Text()); len(groups) > 1 {
			ports = append(ports, groups[1])
		}
	}
	return ports[0], ports[1]
}

func clean(got interface{}) interface{} {
	switch v := got.(type) {
	case TypePFLAA:
		v.BaseSentence = nmea.BaseSentence{}
		return v
	case TypePFLAU:
		v.BaseSentence = nmea.BaseSentence{}
		return v
	case TypePGRMZ:
		v.BaseSentence = nmea.BaseSentence{}
		return v
	default:
		return got
	}
}
