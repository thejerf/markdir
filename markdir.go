package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

var (
	bind        = flag.String("bind", "127.0.0.1:8080", "port to run the server on")
	contentRoot = flag.String("root", ".", "markdown files root dir")
	is_packr    = true
)

type httpHandler struct {
	contentRoot http.Dir
	fileServer  http.Handler
}

func NewDefaultHandler(root http.Dir) http.Handler {
	return &httpHandler{
		contentRoot: root,
		fileServer:  http.FileServer(root),
	}
}

func getTemplate(tmplName string, debug bool) (*template.Template, error) {
	if debug {
		name := filepath.Join("templates", tmplName)
		return template.ParseFiles(name)
	} else {
		templateBox := packr.New("template", "./templates")
		baseHTML, err := templateBox.FindString(tmplName)
		if err != nil {
			return nil, err
		}
		return template.New("base").Parse(baseHTML)
	}
}

func (r *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	upath := req.URL.Path

	isMarkdownFile := strings.HasSuffix(upath, ".md") || strings.HasSuffix(upath, ".markdown")
	if !isMarkdownFile {
		var exist bool
		upath, exist = getREADMEPath(upath, string(r.contentRoot))
		if !exist {
			r.fileServer.ServeHTTP(w, req)
			return
		}
	}

	input, err := ioutil.ReadFile(upath)
	if err != nil {
		msg := fmt.Sprintf("Couldn't read path %s: %v", req.URL.Path, err)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	output := MarkDown(input)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	outputTemplate, err := getTemplate("base.html", is_packr)
	if err != nil {
		msg := fmt.Sprintf("err: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	args := map[string]interface{}{
		"Path": req.URL.Path,
		"Body": template.HTML(string(output)),
	}
	err = outputTemplate.Execute(w, args)
	if err != nil {
		msg := fmt.Sprintf("err: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

}

func getREADMEPath(upath string, contentRoot string) (string, bool) {
	for _, indexPage := range []string{
		"README.md", "README.markdown",
	} {
		indexName := filepath.Join(contentRoot, upath, indexPage)
		if _, err := os.Stat(indexName); err != nil {
			if os.IsNotExist(err) {
				return "", false
			}
		}

		return indexName, true
	}

	return "", false
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()
	root, err := GetContentRoot(*contentRoot)
	if err != nil {
		log.Fatal(err)
	}

	markdownDir := http.Dir(root)
	defaultHandler := NewDefaultHandler(markdownDir)

	mux := http.NewServeMux()
	var staticHandler http.Handler
	if is_packr {
		staticHandler = http.FileServer(http.Dir("static"))
	} else {
		staticBox := packr.New("static", "./static")
		staticHandler = http.FileServer(staticBox)
	}
	mux.Handle("/favicon.ico", staticHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("/", defaultHandler)

	fmt.Printf("Serving %v on %v\n", *contentRoot, *bind)
	log.Fatal(http.ListenAndServe(*bind, logRequest(mux)))
}
