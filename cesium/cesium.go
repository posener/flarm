package cesium

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/posener/goaction/log"
)

type Config struct {
	// Token for Censium service.
	Token string
	// Path for Census web templates.
	Path string
	// KeepAlive is the number of seconds that a FLARM entry that was not update stays visible.
	LiveTime int
	// AltFix paints the objects in this alt diff.
	AltFix int
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
	tmpl, err := template.New("cesium").ParseGlob(filepath.Join(cfg.Path, "script.js"))
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(cfg.Path)))
	mux.Handle("/script.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "script.js", cfg)
		if err != nil {
			log.Printf("Error executing template: %s", err)
		}
	}))
	return mux, nil
}

type Cesium struct {
}
