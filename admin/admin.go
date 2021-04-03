package admin

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/posener/googleauth"
)

type Config struct {
	AllowedEmails []string
}

//go:embed admin.html.gotmpl
var page []byte

func New(cfg Config, path string, data interface{}, reset func()) (*Admin, error) {
	tmpl, err := template.New("admin.html").Parse(string(page))
	if err != nil {
		return nil, err
	}

	if path == "" {
		path = filepath.Join(os.TempDir(), "flarm-config.json")
	}

	allowed := make(map[string]bool, len(cfg.AllowedEmails))
	for _, e := range cfg.AllowedEmails {
		allowed[e] = true
	}

	// If no allowed users were defined, authorization is disabled.
	if len(allowed) == 0 {
		log.Println("Authorization is disabled!")
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return &Admin{
		tmpl:    tmpl,
		data:    string(jsonData),
		path:    path,
		reset:   reset,
		allowed: allowed,
		cfg:     cfg,
	}, nil
}

type Admin struct {
	tmpl    *template.Template
	data    string
	path    string
	reset   func()
	allowed map[string]bool
	cfg     Config
}

func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	creds := googleauth.User(r.Context())
	if len(a.allowed) > 0 && !a.allowed[creds.Email] {
		http.Error(w, fmt.Sprintf("User %s (%s) not allowed", creds.Name, creds.Email), http.StatusForbidden)
		return
	}

	log.Printf("User logged in: %s (%s)", creds.Name, creds.Email)

	switch r.Method {
	case http.MethodPost:
		defer http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		err := r.ParseForm()
		if err != nil {
			log.Printf("Failed parsing form: %s", err)
			return
		}

		// Check reset mode.
		switch m := mode(r.Form); m {
		case "reset":
			log.Println("Requested server reset...")
			a.reset()
		case "update":
			log.Println("Requested config update...")

			data := r.Form.Get("data")
			var v interface{}
			err := json.Unmarshal([]byte(data), &v)
			if err != nil {
				log.Printf("Failed unmarshaling %s: %s", data, err)
				http.Error(w, fmt.Sprintf("Invalid json data: %s", err), http.StatusBadRequest)
				return
			}
			formattedData, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				log.Printf("Failed marshaling data %+v: %s", v, err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			log.Printf("Preparing backup...")
			err = backup(a.path)
			if err != nil {
				log.Printf("Failed preparing backup: %s", err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			log.Printf("Writing new config: \n\n %s\n\n", string(data))
			err = os.WriteFile(a.path, formattedData, 0)
			if err != nil {
				log.Printf("Failed writing config %s: %s", a.path, err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			log.Println("Resetting server...")
			a.reset()
		default:
			log.Printf("Admin got unknown mode: %s", m)
		}
	case http.MethodGet:
		err := a.tmpl.Execute(w, a.data)
		if err != nil {
			log.Printf("Failed executing template: %s", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func mode(v url.Values) string {
	if len(v["mode"]) == 0 {
		return ""
	}
	return v["mode"][0]
}

func backup(path string) error {
	backupPath := path + ".bck"
	dst, err := os.Create(backupPath)
	if err != nil {
		log.Printf("Failed creating backup file: %s", err)
	}
	defer dst.Close()

	src, err := os.Open(path)
	if err != nil {
		log.Printf("Failed creating backup file: %s", err)
	}
	defer src.Close()

	_, err = io.Copy(dst, src)
	return err
}
