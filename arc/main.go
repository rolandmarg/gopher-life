package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	filename := flag.String("file", "story.json", "the json file for story")
	flag.Parse()
	fmt.Printf("Using the story in %s\n", *filename)

	arc, err := jsonStory(*filename)
	if err != nil {
		panic(err)
	}

	h := newHandler(arc)
	log.Fatal(http.ListenAndServe(":8080", h))
}

type story map[string]chapter

type chapter struct {
	Title      string   `json:"title"`
	Paragraphs []string `json:"story"`
	Options    []option `json:"options"`
}

type option struct {
	Text    string `json:"text"`
	Chapter string `json:"arc"`
}

func newHandler(s story) http.Handler {
	return handler{s}
}

type handler struct {
	s story
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimSpace(r.URL.Path)
	s, ok := h.s[p[1:]]
	if p == "/" || p == "/intro" || !ok {
		err := tpl.Execute(w, h.s["intro"])
		if err != nil {
			panic(err)
		}
	} else {
		err := tpl.Execute(w, s)
		if err != nil {
			panic(err)
		}
	}
}

func jsonStory(file string) (story, error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	d := json.NewDecoder(f)
	var arc story
	if err := d.Decode(&arc); err != nil {
		return nil, err
	}

	return arc, nil
}

func init() {
	tpl = template.Must(template.New("").Parse(storyTmpl))
}

var tpl *template.Template

var storyTmpl = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Choose your own adventure</title>
  </head>
  <body>
    <h1>{{.Title}}</h1>
    {{range .Paragraphs}}
    <p>{{.}}</p>
    {{end}}
    <ul>
      {{range .Options}}
      <li><a href="/{{.Chapter}}">{{.Text}}</a></li>
      {{end}}
    </ul>
  </body>
</html>
`
