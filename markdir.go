package main

import (
	"errors"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/russross/blackfriday/v2"
)

const (
	defaultBind = "127.0.0.1:19000"
)

var bind string
var showVersion bool

func init() {
	flag.StringVar(&bind, "bind", defaultBind, "port to run the server on")

	flag.BoolVar(&showVersion, "version", false, "show the version of markdir and exit")
	flag.BoolVar(&showVersion, "v", false, "")
}

func main() {
	flag.Parse()

	if showVersion {
		info, _ := debug.ReadBuildInfo()
		log.Println(info.Main.Path, info.Main.Version)
		os.Exit(0)
	}

	httpdir := http.Dir(".")
	handler := renderer{httpdir, http.FileServer(httpdir)}

	log.Println("Serving on http://" + bind)
	log.Fatal(http.ListenAndServe(bind, handler))
}

var outputTemplate = template.Must(template.New("base").Parse(`
<html>
  <head>
    <title>{{ .Path }}</title>
  </head>
  <body>
    {{ .Body }}
  </body>
</html>
`))

type renderer struct {
	d http.Dir
	h http.Handler
}

func (r renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !strings.HasSuffix(req.URL.Path, ".md") && !strings.HasSuffix(req.URL.Path, "/guide") {
		r.h.ServeHTTP(rw, req)
		return
	}

	// net/http is already running a path.Clean on the req.URL.Path,
	// so this is not a directory traversal, at least by my testing
	var pathErr *os.PathError
	input, err := ioutil.ReadFile("." + req.URL.Path)
	if errors.As(err, &pathErr) {
		http.Error(rw, http.StatusText(http.StatusNotFound)+": "+req.URL.Path, http.StatusNotFound)
		log.Printf("file not found: %s", err)
		return
	}

	if err != nil {
		http.Error(rw, "Internal Server Error: "+err.Error(), 500)
		log.Printf("Couldn't read path %s: %v (%T)", req.URL.Path, err, err)
		return
	}

	output := blackfriday.Run(input)

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	outputTemplate.Execute(rw, struct {
		Path string
		Body template.HTML
	}{
		Path: req.URL.Path,
		Body: template.HTML(string(output)),
	})

}
