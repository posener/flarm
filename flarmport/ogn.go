package flarmport

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"time"
)

// var pattern = regexp.MustCompile(`
//     (?P<pps_offset>\d\.\d+)sec:(?P<frequency>\d+\.\d+)MHz:\s+
//     (?P<aircraft_type>\d):(?P<address_type>\d):(?P<address>[A-F0-9]{6})\s
//     (?P<timestamp>\d{6}):\s
//     \[\s*(?P<latitude>[+-]\d+\.\d+),\s*(?P<longitude>[+-]\d+\.\d+)\]deg\s*
//     (?P<altitude>\d+)m\s*
//     (?P<climb_rate>[+-]\d+\.\d+)m/s\s*
//     (?P<ground_speed>\d+\.\d+)m/s\s*
//     (?P<track>\d+\.\d+)deg\s*
//     (?P<turn_rate>[+-]\d+\.\d+)deg/sec\s*
//     (?P<magic_number>\d+)\s*
//     (?P<gps_status>[0-9x]+)m\s*
//     (?P<channel>\d+)(?P<flarm_timeslot>[f_])(?P<ogn_timeslot>[o_])\s*
//     (?P<frequency_offset>[+-]\d+\.\d+)kHz\s*
//     (?P<decode_quality>\d+\.\d+)/(?P<signal_quality>\d+\.\d+)dB/(?P<demodulator_type>\d+)\s+
//     (?P<error_count>\d+)e\s*
//     (?P<distance>\d+\.\d+)km\s*
//     (?P<bearing>\d+\.\d+)deg\s*
//     (?P<phi>[+-]\d+\.\d+)deg\s*
//     (?P<multichannel>\+)?\s*
//     \?\s*
//     R?\s*
//     (B(?P<baro_altitude>\d+))?
// 	`)

// OGN is a connection to a OGN port.
// Specification at: http://wiki.glidernet.org/wiki:manual-installation-guide.
type OGN struct {
	scanner *bufio.Scanner
	io.Closer
	station StationInfo
}

func OpenOGN(addr string, station StationInfo) (*OGN, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip address: %s", host)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number %q: %v", port, err)
	}
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip, Port: p})
	if err != nil {
		return nil, fmt.Errorf("failed connecting to ogn: %v", err)
	}
	// Create a scanner that splits on CR.
	s := bufio.NewScanner(conn)

	if station.TimeZone == nil {
		station.TimeZone = defaultTimezone
	}

	return &OGN{
		scanner: s,
		Closer:  conn,
		station: station,
	}, nil
}

// Range iterates and parses data from the serial connection. It exists when the port is closed.
func (o *OGN) Range(ctx context.Context, f func(Data)) error {
	for ctx.Err() == nil {
		value, ok := o.next()
		if !ok {
			return nil
		}
		if value != nil && ctx.Err() == nil {
			f(*value)
		}
	}
	return ctx.Err()
}

var pattern = regexp.MustCompile(`(\d+\.\d+)sec:(\d+\.\d+)MHz:\s+(\d+):(\d+):([A-F0-9]+)\s+(\d+):\s+\[\s*([+-]\d+\.\d+),\s*([+-]\d+\.\d+)\]deg\s+(\d+)m\s+([+-]\d+\.\d+)m\/s\s+(\d+.\d+)m\/s\s+(\d+\.\d+)deg\s+([+-]\d+\.\d+)deg`)

// next used by Range and exist for testing purposes.
func (o *OGN) next() (*Data, bool) {
	if !o.scanner.Scan() {
		// Stop scanning.
		return nil, false
	}
	matches := pattern.FindStringSubmatch(o.scanner.Text())
	if len(matches) < 12 {
		return nil, false
	}
	tp := aircraftType(matches[3])
	name := o.station.MapID(matches[5])
	lat, _ := strconv.ParseFloat(matches[7], 64)
	long, _ := strconv.ParseFloat(matches[8], 64)
	alt, _ := strconv.ParseFloat(matches[9], 64)
	climb, _ := strconv.ParseFloat(matches[10], 64)
	gs, _ := strconv.ParseFloat(matches[11], 64)
	dir, _ := strconv.ParseFloat(matches[12], 64)
	tr, _ := strconv.ParseFloat(matches[13], 64)

	return &Data{
		Type:        tp,
		Name:        name,
		Lat:         lat,
		Long:        long,
		Alt:         alt,
		Climb:       climb,
		GroundSpeed: int64(gs),
		Dir:         int(dir),
		TurnRate:    tr,
		Time:        time.Now().In(o.station.TimeZone),
	}, true
}
