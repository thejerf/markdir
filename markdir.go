package main

import (
	"flag"
	"fmt"
	"html"
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

	dir, _ := os.Getwd()
	fmt.Println("Serving", dir)

	httpdir := http.Dir(dir)
	handler := Renderer{httpdir, http.FileServer(httpdir)}

	log.Fatal(http.ListenAndServe(*bind, handler))
}

type Renderer struct {
	d http.Dir
	h http.Handler
}

func (r Renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path, "|--")
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

		_, _ = rw.Write([]byte("<html><head><title>" + html.EscapeString(req.URL.Path) + "</title></head><body>"))
		_, _ = rw.Write(output)
		_, _ = rw.Write([]byte("</body></html>"))
		return
	}

	r.h.ServeHTTP(rw, req)
}
