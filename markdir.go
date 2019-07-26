package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

var bind = flag.String("bind", "127.0.0.1:8080", "port to run the server on")
var contentRoot = flag.String("root", "~/vimwiki", "markdown files root dir")

var (
	TEMPLATE_DIR   string
	STATIC_DIR     string
)

func init() {
	CodeRoot, err := GetCodeRoot()
	if err != nil {
		log.Fatalln(err)
	}
	TEMPLATE_DIR = filepath.Join(CodeRoot, "templates")
	STATIC_DIR = filepath.Join(CodeRoot, "static")

	path := Expand(*contentRoot)
	contentRoot = &path

	fmt.Printf("CodeRoot: %v\nContentRoot: %v\n---\n", CodeRoot, *contentRoot)
}

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

func (r *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	upath := req.URL.Path

	isMarkdownFile := strings.HasSuffix(upath, ".md") || strings.HasSuffix(upath, ".markdown")
	if !isMarkdownFile {
		var exist bool
		upath, exist = getREADMEPath(w, req, r.contentRoot, upath)
		if !exist {
			r.fileServer.ServeHTTP(w, req)
			return
		}
	}

	path := filepath.Join(string(r.contentRoot), upath)
	input, err := ioutil.ReadFile(path)
	if err != nil {
		msg := fmt.Sprintf("Couldn't read path %s: %v", req.URL.Path, err)
		http.Error(w, msg, http.StatusNotFound)
		return
	}
	output := blackfriday.Run(input)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	outputTemplate, err := template.ParseFiles(filepath.Join(TEMPLATE_DIR, "base.html"))
	if err != nil {
		msg := fmt.Sprintf("err: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	err = outputTemplate.Execute(w, struct {
		Path string
		Body template.HTML
	}{
		Path: req.URL.Path,
		Body: template.HTML(string(output)),
	})
	if err != nil {
		msg := fmt.Sprintf("err: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

}

func getREADMEPath(w http.ResponseWriter, r *http.Request, fs http.FileSystem, upath string) (string, bool) {
	const indexPage = "/README.md"
	index := strings.TrimSuffix(upath, "/") + indexPage
	f, err := fs.Open(index)
	if err != nil {
		return "", false
	}

	defer f.Close()
	_, err = f.Stat()
	if err != nil {
		return "", false
	}

	return index, true
}

func main() {
	flag.Parse()

	markdownDir := http.Dir(*contentRoot)
	defaultHandler := NewDefaultHandler(markdownDir)

	mux := http.NewServeMux()
	staticHandler := http.FileServer(http.Dir(STATIC_DIR))
	mux.Handle("/favicon.ico", staticHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("/", defaultHandler)

	fmt.Println("Serving on", *bind)
	log.Fatal(http.ListenAndServe(*bind, mux))
}
