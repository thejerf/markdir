package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	astisub "github.com/asticode/go-astisub"
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
	lastPeriod := strings.LastIndex(req.URL.Path, ".")
	if lastPeriod == -1 {
		r.h.ServeHTTP(rw, req)
		return
	}
	suffix := req.URL.Path[lastPeriod+1:]

	switch suffix {
	case "md":
		r.ServeMarkdown(rw, req)

	case "srt", "ssa", "ass", "st1", "ts", "ttml", "vtt":
		r.ServeSubtitle(rw, req)

	default:
		r.h.ServeHTTP(rw, req)
	}
}

func (r renderer) ServeMarkdown(rw http.ResponseWriter, req *http.Request) {
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

func (r renderer) ServeSubtitle(rw http.ResponseWriter, req *http.Request) {
	parsed, err := astisub.OpenFile("." + req.URL.Path)
	if err != nil {
		http.Error(rw, "Internal Server Error", 500)
		log.Fatalf("Couldn't open path %s: %v", req.URL.Path, err)
	}

	b := []byte("<html><body>")

	for _, item := range parsed.Items {
		b = append(b, []byte("<p><b>")...)
		b = append(b, []byte(item.StartAt.String())...)
		b = append(b, []byte("</b>: ")...)
		for _, line := range item.Lines {
			for _, item := range line.Items {
				b = append(b, []byte(item.Text)...)
				b = append(b, 32)
			}
		}
		b = append(b, []byte("</p>")...)
	}

	b = append(b, []byte("</body></html>")...)

	_, _ = rw.Write(b)
}
