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

	"github.com/russross/blackfriday"
)

var bind = flag.String("bind", "127.0.0.1:8080", "port to run the server on")
var (
	CWD            string
	TEMPLATE_DIR   string
	STATIC_DIR     string
	ContentRootDir string
)

func init() {
	CWD, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
	TEMPLATE_DIR = filepath.Join(CWD, "templates")
	STATIC_DIR = filepath.Join(CWD, "static")
	ContentRootDir = "/Users/michaeltsui/vimwiki"
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
		var isServed bool
		upath, isServed = serveREADME(w, req, r.contentRoot, upath)
		if !isServed {
			r.fileServer.ServeHTTP(w, req)
			return
		}
	}

	path := filepath.Join(ContentRootDir, upath)
	input, err := ioutil.ReadFile(path)
	if err != nil {
		msg := fmt.Sprintf("Couldn't read path %s: %v", req.URL.Path, err)
		http.Error(w, msg, http.StatusNotFound)
		return
	}
	output := blackfriday.Run(input)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	outputTemplate := template.Must(template.ParseFiles(filepath.Join(TEMPLATE_DIR, "base.html")))
	err = outputTemplate.Execute(w, struct {
		Path string
		Body template.HTML
	}{
		Path: req.URL.Path,
		Body: template.HTML(string(output)),
	})

	if err != nil {
		log.Fatalln(err)
	}

}

func serveREADME(w http.ResponseWriter, r *http.Request, fs http.FileSystem, upath string) (string, bool) {
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

	markdownDir := http.Dir(ContentRootDir)
	defaultHandler := NewDefaultHandler(markdownDir)

	mux := http.NewServeMux()
	staticHandler := http.FileServer(http.Dir(STATIC_DIR))
	mux.Handle("/favicon.ico", staticHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("/", defaultHandler)

	fmt.Println("Serving on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(*bind, mux))
}
