package app

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	app = GetApp()
	session := getSession(r)
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == app.Config.AdminUsername && password == app.Config.AdminPassword {
			session.loggedIn = true
			http.Redirect(w, r, session.prevPage, http.StatusFound)
			return
		}
	}
	if session.loggedIn {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	http.ServeFile(w, r, "web/templates/login.go.html")
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				html.WithLineNumbers(true),
			),
		)),
	)

	vars := mux.Vars(r)
	s, err := readMarkdownFile(vars["fileName"])
	if errors.Is(err, os.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "file not found")
	}

	var buf bytes.Buffer
	md.Convert(s, &buf)

	w.WriteHeader(http.StatusOK)
	tmpl, _ := template.ParseFiles("web/templates/view.go.html")

	tmpl.Execute(w, MarkdownHTML{template.HTML(buf.String())})
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
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
			http.Redirect(w, r, fmt.Sprintf("/view/%s", fileName), http.StatusSeeOther)
			return
		}
	}

	s, err := readMarkdownFile(vars["fileName"])
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		internalServerError(w, r)
		return
	}

	tmpl, _ := template.ParseFiles("web/templates/edit.go.html")

	tmpl.Execute(w, Markdown{string(s)})
}

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	fileNames := getAllMdFiles()

	tmpl, _ := template.ParseFiles("web/templates/home.go.html")
	tmpl.Execute(w, fileNames)
}

func getAllMdFiles() []string {
	var mdFiles []string = []string{}
	files, _ := ioutil.ReadDir("./markdown")
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") && !file.IsDir() {
			mdFiles = append(mdFiles, strings.TrimSuffix(file.Name(), ".md"))
		}
	}

	return mdFiles
}

func readMarkdownFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("./markdown/%s.md", fileName))
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "internal server error")
}
