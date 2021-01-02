// Reads and publishes FLARM data.
//
// Get available ports: `dmesg`). If FLARM is connected through serial, it will be on
// `/dev/ttyS[#]`. If connected through USB, it will be on `/dev/ttyUSB[#]`.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/posener/ctxutil"
	"github.com/posener/flarm/cesium"
	"github.com/posener/flarm/flarmport"
	"github.com/posener/flarm/process"
	"github.com/posener/wsbeam"
)

var (
	port       = flag.String("port", "", "Serial port path.")
	addr       = flag.String("addr", ":8080", "Address for HTTP serving.")
	configPath = flag.String("config", "config.json", "Configuration")
)

var cfg struct {
	Location struct {
		Lat  float64
		Long float64
		Alt  float64
	}
	TimeZone string
	// FlarmMap is mapping from FLARM ID to aircraft call name.
	FlarmMap map[string]string
	Cesium   cesium.Config
}

func main() {
	flag.Parse()
	ctx := ctxutil.Interrupt()

	if *port == "" {
		log.Fatalf("Usage: 'port' must be provided.")
	}

	loadConfig()

	location := time.UTC
	if tz := cfg.TimeZone; tz != "" {
		var err error
		location, err = time.LoadLocation(tz)
		if err != nil {
			log.Fatalf("Invalid timezone value %q: %s", tz, err)
		}
	}

	flarm, err := flarmport.Open(*port)
	if err != nil {
		log.Fatal(err)
	}
	defer flarm.Close()

	conns := wsbeam.New()
	cesium, err := cesium.New(cfg.Cesium)
	if err != nil {
		log.Fatalf("Failed loading cesium server: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/ws", conns)
	mux.Handle("/", cesium)
	srv := &http.Server{Addr: *addr, Handler: mux}

	go func() {
		log.Printf("Serving on %s", *addr)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed serving: %s", err)
		}
	}()

	p := process.Processor{
		Lat:      cfg.Location.Lat,
		Long:     cfg.Location.Long,
		Alt:      cfg.Location.Alt,
		IDMap:    cfg.FlarmMap,
		TimeZone: location,
	}

	go func() {
		log.Println("Start reading port...")
		for flarm.Next() {
			if err := flarm.Err(); err != nil {
				log.Printf("Unknown format: %v", err)
			}
			entry := p.Process(flarm.Value())
			if entry != nil {
				conns.Send(entry)
			}
		}
	}()

	<-ctx.Done()

	// Shutdown until killed.
	srv.Shutdown(ctxutil.WithSignal(context.Background(), os.Kill))
}

func loadConfig() {
	b, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed reading config: %s", err)
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalf("Failed parsing config: %s", err)
	}
}
