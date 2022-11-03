package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/gorilla/mux"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

type Markdown struct {
	Text string
}
type MarkdownHTML struct {
	HTML template.HTML
}

func readFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("./markdown/%s.md", fileName))
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "internal server error")
}

func notFoundError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "file not found")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				html.WithLineNumbers(true),
			),
		)),
	)

	vars := mux.Vars(r)
	s, err := readFile(vars["fileName"])
	if errors.Is(err, os.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "file not found")
	}

	var buf bytes.Buffer
	md.Convert(s, &buf)

	w.WriteHeader(http.StatusOK)
	tmpl, _ := template.ParseFiles("view.go.html")

	tmpl.Execute(w, MarkdownHTML{template.HTML(buf.String())})
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["fileName"]

	if r.Method == "POST" {
		f := r.FormValue("markdown")
		if f != "" {
			err := os.WriteFile(fmt.Sprintf("markdown/%s.md", fileName), []byte(f), 0644)
			if err != nil {
				internalServerError(w, r)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/view/%s", fileName), http.StatusMovedPermanently)
			return
		}
	}

	s, err := readFile(vars["fileName"])
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		internalServerError(w, r)
		return
	}

	tmpl, _ := template.ParseFiles("edit.go.html")

	tmpl.Execute(w, Markdown{string(s)})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/view/{fileName}", viewHandler)
	r.HandleFunc("/{fileName}", editHandler)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))
}
