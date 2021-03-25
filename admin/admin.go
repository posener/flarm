package admin

import (
	"bytes"
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
	"reflect"
	"sort"
	"strconv"
	"strings"
)

//go:embed admin.html.gotmpl
var page []byte

func New(path string, cfg interface{}, reset func()) (*Admin, error) {
	tmpl, err := template.New("admin.html").Parse(string(page))
	if err != nil {
		return nil, err
	}

	cfg, err = convertToMap(cfg)
	if err != nil {
		return nil, err
	}

	entries := flatten(cfg)

	var html bytes.Buffer
	err = tmpl.Execute(&html, entries)
	if err != nil {
		return nil, err
	}

	if path == "" {
		path = filepath.Join(os.TempDir(), "flarm-config.json")
	}

	return &Admin{
		path:    path,
		html:    html.Bytes(),
		reset:   reset,
		entries: entries,
	}, nil
}

type Admin struct {
	path    string
	html    []byte
	reset   func()
	entries []entry
}

func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			entries := loadFormEntries(a.entries, r.Form)
			c := combine(entries)
			data, err := json.MarshalIndent(c, "", "  ")
			if err != nil {
				log.Printf("Failed marshaling %+v: %s", c, err)
				return
			}
			log.Printf("Preparing backup...")
			err = backup(a.path)
			if err != nil {
				log.Printf("Failed preparing backup: %s", err)
				return
			}

			log.Printf("Writing new config: \n\n %s\n\n", string(data))
			err = os.WriteFile(a.path, data, 0)
			if err != nil {
				log.Printf("Failed writing config %s: %s", a.path, err)
				return
			}
			log.Println("Resetting server...")
			a.reset()
		default:
			log.Printf("Admin got unknown mode: %s", m)
		}
	case http.MethodGet:
		w.Write(a.html)
	}
}

type entry struct {
	Key   string
	Value string
	Type  string // HTML input type.
	Kind  reflect.Kind
}

func flatten(v interface{}) []entry {
	var entries []entry
	flattenPart(v, "", &entries)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	return entries
}

func flattenPart(v interface{}, prefix string, entries *[]entry) {
	if v == nil {
		return
	}
	e := entry{
		Key:   strings.TrimLeft(prefix, "."),
		Value: fmt.Sprintf("%v", v),
	}
	switch t := v.(type) {
	case string:
		e.Type = "text"
		e.Kind = reflect.String
		*entries = append(*entries, e)
	case int:
		e.Type = "number"
		e.Kind = reflect.Int
		*entries = append(*entries, e)
	case float64:
		e.Type = "number"
		e.Kind = reflect.Float64
		*entries = append(*entries, e)
	case map[string]interface{}:
		for key, val := range t {
			flattenPart(val, prefix+"."+key, entries)
		}
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

func loadFormEntries(entries []entry, values url.Values) []entry {
	set := map[string]entry{}
	for _, e := range entries {
		if v := values.Get(e.Key); v != "" {
			e.Value = v
		}
		set[e.Key] = e
	}
	ret := []entry{}
	for _, e := range set {
		ret = append(ret, e)
	}
	return ret
}

func combine(entries []entry) interface{} {
	v := map[string]interface{}{}
	for _, entry := range entries {
		parts := strings.Split(entry.Key, ".")
		u := v
		for _, part := range parts[:len(parts)-1] {
			if u[part] == nil {
				u[part] = map[string]interface{}{}
			}
			u = u[part].(map[string]interface{})
		}
		var err error
		u[parts[len(parts)-1]], err = convert(entry.Kind, entry.Value)
		if err != nil {
			panic(fmt.Sprintf("Failed key %s with value %s: %s", entry.Key, entry.Value, err))
		}
	}
	return v
}

func convert(k reflect.Kind, v string) (interface{}, error) {
	switch k {
	case reflect.String:
		return v, nil
	case reflect.Int:
		return strconv.Atoi(v)
	case reflect.Float64:
		return strconv.ParseFloat(v, 64)
	default:
		return nil, fmt.Errorf("unexpected kind: %s", k)
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

// convertToMap converts the given value to a Go map by marshaling and unmarshaling.
func convertToMap(v interface{}) (interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return v, err
}
