package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/russross/blackfriday"
)

var bind = flag.String("bind", "127.0.0.1:19000", "port to run the server on")

func main() {
	flag.Parse()

	httpdir := http.Dir(".")
	handler := renderer{httpdir, http.FileServer(httpdir)}

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe(*bind, handler))
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
	if !strings.HasSuffix(req.URL.Path, ".md") {
		r.h.ServeHTTP(rw, req)
		return
	}

	// net/http is already running a path.Clean on the req.URL.Path,
	// so this is not a directory traversal, at least by my testing
	input, err := ioutil.ReadFile("." + req.URL.Path)
	if err != nil {
		http.Error(rw, "Internal Server Error", 500)
		log.Fatalf("Couldn't read path %s: %v", req.URL.Path, err)
	}
	output := blackfriday.MarkdownCommon(input)

	rw.Header().Set("Content-Type", "text/html")

	outputTemplate.Execute(rw, struct {
		Path string
		Body template.HTML
	}{
		Path: req.URL.Path,
		Body: template.HTML(string(output)),
	})

}
