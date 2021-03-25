package cesium

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
)

//go:embed web/*
var embedFS embed.FS

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
	// Units to show. "metric", "imperial" or "mixed"..
	Units string
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
	if cfg.Units != "metric" && cfg.Units != "imperial" && cfg.Units != "mixed" {
		return nil, fmt.Errorf("units want: [metric,imperial,mixed], got: %s", cfg.Units)
	}
	mux := http.NewServeMux()

	// Create a handler that holds the unmodified content.
	unmodifiedFS, err := fs.Sub(embedFS, "web")
	if err != nil {
		return nil, fmt.Errorf("no subdir 'web' in filesystem")
	}
	err = mount(cfg, mux, "/nomod/", unmodifiedFS)
	if err != nil {
		return nil, fmt.Errorf("creating unmodified handler: %s", err)
	}

	// Create a handler where the static content can be modified according to a given on-disk
	// content, according to the configured cfg.Path.
	err = mount(cfg, mux, "/", unionFS{os.DirFS(cfg.Path), unmodifiedFS})
	if err != nil {
		return nil, fmt.Errorf("creating modified handler: %s", err)
	}
	return mux, nil
}

// mount returns a web http mount for the given filesystem, updating the script.js according to
// the given config.
func mount(cfg Config, mux *http.ServeMux, prefix string, fsys fs.FS) error {
	templates, err := template.ParseFS(fsys, "script.js")
	if err != nil {
		return err
	}

	// Format the script according to the config.
	script := bytes.Buffer{}
	err = templates.ExecuteTemplate(&script, "script.js", cfg)
	if err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}
	scriptBytes := script.Bytes()

	mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.FS(fsys))))
	mux.Handle(path.Join(prefix, "script.js"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(scriptBytes)
		if err != nil {
			log.Printf("Failed writing script: %s", err)
		}
	}))
	return nil
}

// unionFS returns the union of multiple filesystems. When asked for a file it checks each
// filesystem in the given order until one of them does not return ErrNotExist. In that case the
// file and error are returned. If all the filesystems returned an ErrNotExist, it will be also
// returned to the user.
type unionFS []fs.FS

func (u unionFS) Open(name string) (fs.File, error) {
	for _, i := range u {
		f, err := i.Open(name)
		if !errors.Is(err, fs.ErrNotExist) {
			return f, err
		}
	}
	return nil, fs.ErrNotExist
}
