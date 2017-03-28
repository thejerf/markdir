package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/russross/blackfriday"
)

var bind = flag.String("bind", "127.0.0.1:19000", "port to run the server on")

func main() {
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't get the current working dir from the os: " +
			err.Error())
	}

	httpdir := http.Dir(dir)
	handler := Renderer{httpdir, http.FileServer(httpdir)}

	fmt.Println("Serving", dir)
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

type Renderer struct {
	d http.Dir
	h http.Handler
}

func (r Renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, ".md") {
		f, err := r.d.Open(req.URL.Path)
		if err != nil {
			panic(err)
		}
		input, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
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

		return
	}

	r.h.ServeHTTP(rw, req)
}
