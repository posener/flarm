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
	"os/signal"
	"time"

	"github.com/posener/flarm/admin"
	"github.com/posener/flarm/cesium"
	"github.com/posener/flarm/flarmport"
	"github.com/posener/flarm/flarmremote"
	"github.com/posener/flarm/googleauth"
	"github.com/posener/flarm/logger"
	"github.com/posener/flarm/process"
	"github.com/posener/wsbeam"
)

var (
	port       = flag.String("port", "", "Serial port path.")
	remote     = flag.String("remote", "", "Remote flarm server to connect to.")
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
	SSL      struct {
		Cert string
		Key  string
	}
	Log        logger.Config
	Admin      admin.Config
	GoogleAuth googleauth.Config
}

// Common interface for flarmport and flarmremote.
type flarmReader interface {
	// Range iterates over the values received from the flarm.
	Range(func(interface{})) error
	// Close stops reading flarm data.
	Close() error
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	for ctx.Err() == nil {
		serve(ctx)
	}

	<-ctx.Done()
}

func serve(ctx context.Context) {
	// Create cancel handler for this context.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Load config in case it was updated.
	loadConfig()

	location := time.UTC
	if tz := cfg.TimeZone; tz != "" {
		var err error
		location, err = time.LoadLocation(tz)
		if err != nil {
			log.Fatalf("Invalid timezone value %q: %s", tz, err)
		}
	}

	sendLog, err := logger.New(cfg.Log)
	if err != nil {
		log.Fatalf("Failed initializing logger: %s", err)
	}

	var flarm flarmReader
	switch {
	case *port != "" && *remote != "":
		log.Fatal("Usage: can't provide both 'port' and 'remote'.")
	case *port != "":
		flarm, err = flarmport.Open(*port)
	case *remote != "":
		flarm, err = flarmremote.Open(*remote)
	default:
		log.Fatal("Usage: must provide 'port' or 'remote'.")
	}
	if err != nil {
		log.Fatal(err)
	}
	defer flarm.Close()

	conns := wsbeam.New()
	cesium, err := cesium.New(cfg.Cesium)
	if err != nil {
		log.Fatalf("Failed loading cesium server: %s", err)
	}

	adminHandler, err := admin.New(cfg.Admin, *configPath, cfg, cancel)
	if err != nil {
		log.Fatalf("Failed loading admin handler: %s", err)
	}

	auth, err := googleauth.New(ctx, cfg.GoogleAuth)
	if err != nil {
		log.Fatalf("Failed loading auth middleware: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/ws", conns)
	mux.Handle("/", cesium)
	mux.Handle("/admin", http.StripPrefix("/admin", auth.Authenticate(adminHandler)))
	mux.Handle("/auth", auth.RedirectHandler())
	srv := &http.Server{Addr: *addr, Handler: mux}

	go func() {
		log.Printf("Serving on %s", *addr)
		var err error
		if cfg.SSL.Key != "" && cfg.SSL.Cert != "" {
			err = srv.ListenAndServeTLS(cfg.SSL.Cert, cfg.SSL.Key)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil {
			log.Print(err)
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
		log.Println("Start reading flarm data...")
		err := flarm.Range(func(value interface{}) {
			entry := p.Process(value)
			if entry != nil {
				log.Printf("sending %+v", entry)
				sendLog.Log(entry)
				conns.Send(entry)
			}
		})
		if err != nil {
			log.Fatalf("Failed iterating flarm values: %v", err)
		}
	}()

	<-ctx.Done()

	// Gracefully shutdown. Allow 1m for connections to disconnect.
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()
	srv.Shutdown(ctx)
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
