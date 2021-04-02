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
	"fmt"
	"io"

	"github.com/adrianmo/go-nmea"
	"github.com/jacobsa/go-serial/serial"
)

// Port is a connection to a FLARM serial port.
type Port struct {
	scanner *bufio.Scanner
	io.Closer
}

// Open opens a serial connection to a given FLARM port.
func Open(port string, baudRate uint) (*Port, error) {
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

	return &Port{scanner: s, Closer: serial}, nil
}

// Range iterates and parses data from the serial connection. It exists when the port is closed.
func (p *Port) Range(f func(interface{})) error {
	for {
		value, ok := p.next()
		if !ok {
			return nil
		}
		if value != nil {
			f(value)
		}
	}
}

// next used by Range and exist for testing purposes.
func (p *Port) next() (interface{}, bool) {
	if !p.scanner.Scan() {
		return nil, false
	}
	line := p.scanner.Text()
	value, err := nmea.Parse(line)
	if err != nil {
		// Unknown NMEA, ignore...
	}
	return value, true
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
