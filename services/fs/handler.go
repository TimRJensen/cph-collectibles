package main

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	dir string
)

func init() {
	// if v, ok := os.LookupEnv("MODE"); !ok {
	// 	panic("environment variable MODE not set")
	// } else {
	// 	if v == "dev" {
	// 		return
	// 	}
	// }

	// if v, ok := os.LookupEnv("FILE_DIR"); !ok {
	// 	panic("environment variable FILE_DIR not set")
	// } else {
	// 	dir = v
	// }
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := filepath.Clean(r.PathValue("id"))
		name := filepath.Clean(r.PathValue("file"))

		// make path
		cwd, _ := os.Getwd()
		filename, err := filepath.Abs(filepath.Join(cwd, "assets", id, name))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		st, err := os.Stat(filename)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		file, err := os.Open(filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			defer file.Close()
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ctype := mime.TypeByExtension(ext); ctype != "" {
			w.Header().Set("Content-Type", ctype)
		}

		http.ServeContent(w, r, st.Name(), st.ModTime(), file)
	}
}
