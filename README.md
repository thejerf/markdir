markdir
=======

Markdir serves up a directory of files in HTML, rendering Markdown files
into HTML when they are encountered as `.md` files. It's sort of the most
degenerate Wiki you could imagine writing short of simply having static
HTML files.

Markdir started out as a simple web application that did something useful
for me. However, I also noticed when I was done that this was a
reasonable answer to a question I've seen on /r/golang a few times: "Does
anyone have a Go web app that does something non-trivial that I can look at
to see how it all works?"

Yes. This actually works out pretty well. This uses enough features that
it's not a "hello world" app:

  * Loads an external library from github.com.
  * Uses the flag library to determine where to bind.
  * Uses the built-in HTTP server.
  * Uses the built-in html/template code.
    * Including a use of template.HTML to indicate already-HTML-encoded text.
  * Implements a struct-based HTTP handler.
  * Demonstrates both declared and anonymous struct usage.
  * Demonstrates wrapping an existing HTTP handler, that is to say, the
    Markdown server is the very simplest sort of "middleware" you can have,
    manually spelled out without a "framework".

This is not written to be a "tour de force" of Go, nor to show off my own
skills, nor to be the fanciest thing ever... it's written to be a readable,
useful example. It's fewer bytes than this README.md is.

Installation
============

    go get -v -u github.com/thejerf/markdir
    ./markdir -h  # see the default flag help file
    ./markdir     # serves the directory

The `-v` is just to show you what is being installed, since this is a
learning exercise. `-u` says to update if necessary.

Navigate to [http://localhost:19000](http://localhost:19000) by default to
see the server. You may need to find some markdown files to see the program
doing anything useful, though. If you start it up in the directory this
repo clones into, you can go
to [http://localhost:19000/README.md](http://localhost:19000/README.md) to
read this very file through the server.

Security
========

This HTTP server accepts no commands beyond GET commands to read the files
in your directory. In theory, because this only reads files off of the
disk, even if the blackfriday library is vulnerable to something you'd
still have to have disk access to exploit it.

That said, I have this bind to localhost by default on purpose. I'd think
twice about opening this up to the internet, because you should always
think twice about that. If I were going to do that, I'd probably still take
the step of running the output of blackfriday
through [bluemonday](https://github.com/microcosm-cc/bluemonday).

Release History
===============

  * v1.0.1: Internal rename to make this lint-clean by my gometalinter standards.
  * v1.0: Initial release.

