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
	fsys := unionFS{
		os.DirFS(cfg.Path),
		mustSub(embedFS, "web"),
	}

	templates := template.Must(template.ParseFS(fsys, "script.js"))

	script := bytes.Buffer{}
	err := templates.ExecuteTemplate(&script, "script.js", cfg)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(fsys)))
	mux.Handle("/script.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(script.Bytes())
		if err != nil {
			log.Printf("Failed writing script: %s", err)
		}
	}))
	return mux, nil
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

// mustSub returns a filesystem in a subdirectory, or panics if directory does not exist.
func mustSub(f fs.FS, dir string) fs.FS {
	f, err := fs.Sub(f, dir)
	if err != nil {
		panic(fmt.Sprintf("No subdir %q in filesystem.", dir))
	}
	return f
}
