package cesium

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"text/template"
)

//go:embed web/*
var static embed.FS

var templates = template.Must(template.ParseFS(static, "web/script.js"))

type Config struct {
	// Token for Censium service.
	Token string
	// Path for Census web templates.
	Path string
	// AltFix paints the objects in this alt diff.
	AltFix int
	// PathLength is the number of path steps to show, after which the path is deleted.
	PathLength int
	// MinGroundSpeed is the minimum ground speed (in m/s) to show an aircraft.
	MinGroundSpeed float32
	// Start location
	Camera struct {
		Lat     float64
		Long    float64
		Alt     float64
		Heading float64
		Pitch   float64
	}
}

func New(cfg Config) (http.Handler, error) {
	script := bytes.Buffer{}
	err := templates.ExecuteTemplate(&script, "script.js", cfg)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %s", err)
	}

	serveDir, err := fs.Sub(static, "web")
	if err != nil {
		panic("no subdir 'web' in filesystem.")
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(serveDir)))
	mux.Handle("/script.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(script.Bytes())
		if err != nil {
			log.Printf("Failed writing script: %s", err)
		}
	}))
	return mux, nil
}

type Cesium struct {
}
