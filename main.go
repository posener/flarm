// Reads and publishes FLARM data.
//
// Get available ports: `dmesg`). If FLARM is connected through serial, it will be on
// `/dev/ttyS[#]`. If connected through USB, it will be on `/dev/ttyUSB[#]`.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/posener/auth"
	"github.com/posener/flarm/admin"
	"github.com/posener/flarm/cesium"
	"github.com/posener/flarm/flarmport"
	"github.com/posener/flarm/logger"
	"github.com/posener/wsbeam"
	"golang.org/x/crypto/acme/autocert"
)

var (
	port     = flag.String("port", "", "Serial port path to connect to.")
	baudRate = flag.Uint("baud_rate", 57600, "Serial port baud rate.")

	remote = flag.String("remote", "", "Remote flarm server to connect to.")

	ogn = flag.String("ogn", "", "OGN address to connect to")

	addr       = flag.String("addr", ":8082", "Address for HTTP serving.")
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
		Cert        string
		Key         string
		LetsEncrypt struct {
			Enabled      bool
			AllowedHosts []string
			CacheDir     string
		}
	}
	Log        logger.Config
	Admin      admin.Config
	GoogleAuth auth.Config

	FlarmReconnectDelaySec int
}

const defaultFlarmReconnectDelay = time.Second * 3

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

	conns := wsbeam.New()
	cesium, err := cesium.New(cfg.Cesium)
	if err != nil {
		log.Fatalf("Failed loading cesium server: %s", err)
	}

	adminHandler, err := admin.New(cfg.Admin, *configPath, cfg, cancel)
	if err != nil {
		log.Fatalf("Failed loading admin handler: %s", err)
	}

	authHandler, err := auth.New(ctx, cfg.GoogleAuth)
	if err != nil {
		log.Fatalf("Failed loading auth middleware: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/ws", conns)
	mux.Handle("/", cesium)
	mux.Handle("/admin", http.StripPrefix("/admin", authHandler.Authenticate(adminHandler)))
	mux.Handle("/auth", authHandler.RedirectHandler())
	srv := &http.Server{Addr: *addr, Handler: mux}

	go func() {
		log.Printf("Serving on %s", *addr)
		var err error
		switch {
		case cfg.SSL.Key != "" && cfg.SSL.Cert != "":
			err = srv.ListenAndServeTLS(cfg.SSL.Cert, cfg.SSL.Key)
		case cfg.SSL.LetsEncrypt.Enabled:
			cm := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(cfg.SSL.LetsEncrypt.AllowedHosts...),
				Cache:      autocert.DirCache(cfg.SSL.LetsEncrypt.CacheDir),
			}
			srv.TLSConfig = &tls.Config{
				GetCertificate: cm.GetCertificate,
			}
			go func() {
				err := http.ListenAndServe(":80", cm.HTTPHandler(nil))
				if err != nil {
					log.Fatalf("Failed autocert serving: %s", err)
				}
			}()
			err = srv.ListenAndServeTLS("", "")
		default:
			err = srv.ListenAndServe()
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	station := flarmport.StationInfo{
		Lat:      cfg.Location.Lat,
		Long:     cfg.Location.Long,
		Alt:      cfg.Location.Alt,
		IDMap:    cfg.FlarmMap,
		TimeZone: location,
	}

	go func() {
		for ctx.Err() == nil {
			flarm, err := getFlarm(station)
			if err == nil {
				defer flarm.Close()
				log.Println("Start reading flarm data...")
				err = flarm.Range(ctx, func(o flarmport.Data) {
					log.Printf("sending %+v", o)
					sendLog.Log(o)
					conns.Send(o)
				})
			}

			if err != nil {
				log.Printf("Failed iterating flarm values: %v", err)
			}
			// If context was not cancelled, reconnect to flarm.
			if ctx.Err() == nil {
				flarmReconnectDelay := time.Duration(cfg.FlarmReconnectDelaySec) * time.Second
				if flarmReconnectDelay == 0 {
					flarmReconnectDelay = defaultFlarmReconnectDelay
				}
				log.Printf("Will try to reconnect to flarm in %v...", flarmReconnectDelay)
				time.Sleep(flarmReconnectDelay)
			}
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
	cfg.GoogleAuth.Log = log.Printf

	// Check SSL config.
	if ssl := cfg.SSL; ssl.Cert != "" || ssl.Key != "" {
		if ssl.LetsEncrypt.Enabled {
			log.Fatal("Cant use SSL.Cert and SSL.Key with SSL.LetsEncrypt.Enabled.")
		}
		if ssl.Cert == "" || ssl.Key == "" {
			log.Fatalf("When using SSL Cert and Key, both Cert and Key should be set.")
		}
	}
	if letsEncrypt := cfg.SSL.LetsEncrypt; letsEncrypt.Enabled {
		if len(letsEncrypt.AllowedHosts) == 0 {
			log.Fatalf("When LetsEncrypt is enabled, AllowedHosts must be given.")
		}
	}
}

func getFlarm(station flarmport.StationInfo) (flarmport.Reader, error) {
	switch {
	case countInputSelection() > 1:
		log.Fatal("Usage: can't use multiple sources. Must select one of 'port', 'ogn' or 'remote'.")
	case *port != "":
		return flarmport.Open(*port, *baudRate, station)
	case *ogn != "":
		return flarmport.OpenOGN(*ogn, station)
	case *remote != "":
		return flarmport.Remote(*remote)
	}
	return nil, fmt.Errorf("usage: must provide 'port' or 'remote'")
}

func countInputSelection() int {
	s := 0
	if *port != "" {
		s++
	}
	if *ogn != "" {
		s++
	}
	if *remote != "" {
		s++
	}
	return s
}
