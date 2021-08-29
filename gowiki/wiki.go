// https://golang.org/doc/articles/wiki/

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// data structure
type Page struct {
	Title string
	Body  []byte
}

// caching templates
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

// validating title by checking if it is a "normal" string
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// check if a title string has a valid content
// func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
// 	m := validPath.FindStringSubmatch(r.URL.Path)
// 	if m == nil {
// 		http.NotFound(w, r)
// 		return "", errors.New("invalid Page Title")
// 	}
// 	return m[2], nil // The title is the second subexpression.
// }

// closure function to check if a title string has a valid content
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

// write a .txt file to disk with given data
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile("data/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile("data/" + filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// bootstrap the application
// start the server and bind different handlers to each http route
func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
