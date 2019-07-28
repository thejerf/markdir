package main

import (
	"github.com/russross/blackfriday"
)

func getOpts() []blackfriday.Option {
	//TODO: Need to support format `+ [x] some text`
	var opts []blackfriday.Option

	render := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	opts = append(opts, blackfriday.WithRenderer(render))

	return opts
}

func MarkDown(input []byte) []byte {
	opts := getOpts()
	output := blackfriday.Run(input, opts...)

	return output
}
