// Reads and publishes FLARM data.
//
// Get available ports: `dmesg`). If FLARM is connected through serial, it will be on
// `/dev/ttyS[#]`. If connected through USB, it will be on `/dev/ttyUSB[#]`.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/posener/ctxutil"
	"github.com/posener/flarm/connections"
	"github.com/posener/flarm/flarmport"
	"github.com/posener/flarm/process"
)

var (
	port = flag.String("port", "", "Serial port path.")
	addr = flag.String("addr", ":8080", "Address for HTTP serving.")

	lat  = flag.Float64("lat", 32.578190, "Latitude of station.")
	long = flag.Float64("long", 35.178741, "Longitude of station.")
	alt  = flag.Float64("alt", 69, "Altitude of station in meters.")
)

var config struct {
	Lat  float64
	Long float64
	Alt  float64
}

func main() {
	flag.Parse()
	ctx := ctxutil.Interrupt()

	if *port == "" {
		log.Fatalf("Usage: 'port' must be provided.")
	}

	flarm, err := flarmport.Open(*port)
	if err != nil {
		log.Fatal(err)
	}
	defer flarm.Close()

	conns := connections.New()
	srv := &http.Server{Addr: *addr, Handler: conns}

	go func() {
		log.Printf("Serving websocket on ws://%s", *addr)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed serving: %s", err)
		}
	}()

	p := process.Processor{
		Lat:  *lat,
		Long: *long,
		Alt:  *alt,
	}

	go func() {
		log.Println("Start reading port...")
		for flarm.Next() {
			if err := flarm.Err(); err != nil {
				log.Printf("Unknown format: %v", err)
			}
			entry := p.Process(flarm.Value())
			if entry != nil {
				conns.Write(entry)
			}
		}
	}()

	<-ctx.Done()

	// Shutdown until killed.
	srv.Shutdown(ctxutil.WithSignal(context.Background(), os.Kill))
}
