// flarmport a library for connecting and reading from FLARM serial port.
//
// According to the flarm specification, available on:
// http://www.ediatec.ch/pdf/FLARM%20Data%20Port%20Specification%20v7.00.pdf
//
// A usage example:
//
// 	flarm, err := flarmport.Open("/dev/ttyS0")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer flarm.Close()
// 	for flarm.Next() {
// 		if err := flarm.Err(); err != nil {
// 			log.Printf("Unknown format: %v", err)
// 		}
// 		entry := flarm.Value()
// 		if entry != nil {
// 			fmt.Printf("%+v", entry)
// 		}
// 	}
package flarmport

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/jacobsa/go-serial/serial"
)

var defaultTimezone = time.UTC

// StationInfo is information about the station where the flarm is location.
type StationInfo struct {
	// Latitude and longitude coordinates of the station.
	Lat, Long float64
	// Altitude of the station, in meters.
	Alt float64
	// Time zone of the station. Default is set to UTC.
	TimeZone *time.Location
	// Mapping of flarm ID to plane sign.
	IDMap map[string]string
}

func (si StationInfo) MapID(id string) string {
	if mapped := si.IDMap[id]; mapped != "" {
		return mapped
	}
	return id
}

// Port is a connection to a FLARM serial port.
type Port struct {
	scanner *bufio.Scanner
	io.Closer
	station StationInfo
}

// Open opens a serial connection to a given FLARM port.
func Open(port string, baudRate uint, station StationInfo) (*Port, error) {
	serial, err := serial.Open(serial.OpenOptions{
		PortName: port,
		// Baud rate from spec: "The baud rate can be configured by commands described in FLARM
		// configuration specification"
		BaudRate:        baudRate,
		MinimumReadSize: 1,
		StopBits:        1,                  // From spec: "1 stop bit".
		DataBits:        8,                  // From spec: "All ports use 8 data bits".
		ParityMode:      serial.PARITY_NONE, // From spec: "no parity".

	})
	if err != nil {
		return nil, fmt.Errorf("failed open serial port: %v", err)
	}
	// Create a scanner that splits on CR.
	s := bufio.NewScanner(serial)
	s.Split(splitCR)

	if station.TimeZone == nil {
		station.TimeZone = defaultTimezone
	}

	return &Port{
		scanner: s,
		Closer:  serial,
		station: station,
	}, nil
}

// Range iterates and parses data from the serial connection. It exists when the port is closed.
func (p *Port) Range(ctx context.Context, f func(Data)) error {
	for ctx.Err() == nil {
		value, ok := p.next()
		if !ok {
			return nil
		}
		if value != nil && ctx.Err() == nil {
			f(*value)
		}
	}
	return ctx.Err()
}

// next used by Range and exist for testing purposes.
func (p *Port) next() (*Data, bool) {
	if !p.scanner.Scan() {
		// Stop scanning.
		return nil, false
	}
	line := p.scanner.Text()
	value, err := nmea.Parse(line)
	if err != nil {
		// Unknown NMEA, ignore...
		return nil, true
	}

	switch e := value.(type) {
	case TypePFLAA:
		return p.station.processPFLAA(e), true
	}
	return nil, true
}

func splitCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

type Data struct {
	Name      string
	Lat, Long float64 `gorm:"type=float,precision=2"`
	// Direction of airplane (In degrees relative to N)
	Dir int
	// Altitude in m
	Alt float64
	// Ground speed in m/s
	GroundSpeed int64
	// Climb rate in m/s
	Climb float64
	// Change to direction in deg/s
	TurnRate   float64
	Type       string
	Time       time.Time
	AlarmLevel int
}

func (o *Data) TableName() string { return "logs" }

func (s StationInfo) processPFLAA(e TypePFLAA) *Data {
	id := s.MapID(e.ID)
	if id == "" {
		log.Println("Ignoring empty ID entry.")
		return nil
	}
	lat, long := add(s.Lat, s.Long, float64(e.RelativeNorth), float64(e.RelativeEast))
	return &Data{
		Name:        id,
		Lat:         lat,
		Long:        long,
		Dir:         int(e.Track),
		Alt:         s.Alt + float64(e.RelativeVertical),
		Climb:       e.ClimbRate,
		GroundSpeed: e.GroundSpeed,
		Type:        e.AircraftType,
		AlarmLevel:  int(e.AlarmLevel),
		Time:        time.Now().In(s.TimeZone),
	}
}

func add(lat, lon float64, relN, relE float64) (float64, float64) {
	const earthRadius = 6378137

	//Coordinate offsets in radians
	dLat := relN / earthRadius
	dLon := relE / (earthRadius * math.Cos(math.Pi*lat/180.0))

	//OffsetPosition, decimal degrees
	return lat + dLat*180.0/math.Pi,
		lon + dLon*180.0/math.Pi
}
