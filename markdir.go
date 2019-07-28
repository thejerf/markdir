package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

var (
	bind        = flag.String("bind", "127.0.0.1:8080", "port to run the server on")
	contentRoot = flag.String("root", ".", "markdown files root dir")
	isPackr     = true
)

type httpHandler struct {
	contentRoot string
	fileServer  http.Handler
}

func NewDefaultHandler(root string) http.Handler {
	return &httpHandler{
		contentRoot: root,
		fileServer:  http.FileServer(http.Dir(root)),
	}
}

func getTemplate(tmplName string, isPackr bool) (*template.Template, error) {
	if isPackr {
		templateBox := packr.New("template", "./templates")
		baseHTML, err := templateBox.FindString(tmplName)
		if err != nil {
			return nil, err
		}
		return template.New("base").Parse(baseHTML)
	} else {
		name := filepath.Join("templates", tmplName)
		return template.ParseFiles(name)
	}
}

func (r *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var filename string
	upath := req.URL.Path

	isMarkdownFile := strings.HasSuffix(upath, ".md") || strings.HasSuffix(upath, ".markdown")
	if !isMarkdownFile {
		var exist bool
		filename, exist = getREADMEPath(upath, r.contentRoot)
		if !exist {
			r.fileServer.ServeHTTP(w, req)
			return
		}
	} else {
		filename = filepath.Join(r.contentRoot, upath)
	}

	input, err := ioutil.ReadFile(filename)
	if err != nil {
		msg := fmt.Sprintf("Couldn't read path %s: %v", req.URL.Path, err)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	output := MarkDown(input)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	outputTemplate, err := getTemplate("base.html", isPackr)
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
		urlStr := fmt.Sprintf("%v", r.URL)
		path, err := url.QueryUnescape(urlStr)
		if err != nil {
			path = urlStr
		}
		log.Printf("%s %s\n", r.Method, path)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()
	root, err := GetContentRoot(*contentRoot)
	if err != nil {
		log.Fatal(err)
	}

	defaultHandler := NewDefaultHandler(root)

	mux := http.NewServeMux()
	var staticHandler http.Handler
	if isPackr {
		staticBox := packr.New("static", "./static")
		staticHandler = http.FileServer(staticBox)
	} else {
		staticHandler = http.FileServer(http.Dir("static"))
	}
	mux.Handle("/favicon.ico", staticHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("/", defaultHandler)

	fmt.Printf("Serving %v on %v\n", root, *bind)
	log.Fatal(http.ListenAndServe(*bind, logRequest(mux)))
}
