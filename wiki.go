package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// Page is a struct used here
// has Title and Body
type Page struct {
	Title string
	Body  template.HTML
}

const templateDir = "tmpl/"
const dataDir = "data/"

var templates = template.Must(template.ParseFiles(templateDir+"edit.html", templateDir+"view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var pageInterlinkPattern = regexp.MustCompile("\\[([A-Za-z0-9])+\\]")

func (p *Page) save() error {
	filename := dataDir + p.Title + ".txt"
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := dataDir + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
  linksToReplace := pageInterlinkPattern.FindAllString(body, -1)
  if linksToReplace != nil {
    for _, v := range linksToReplace {
      linkTitle := v[1:len(v)-1]
      body = regexp.MustCompile("\\[" + linkTitle + "\\]").ReplaceAllString(body, " <a href=\"/view/" + linkTitle + "\">" + linkTitle + "</a>")
    }
  }
	p := &Page{Title: title, Body: template.HTML(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func redirectToFrontPageHandler(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

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

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid page title")
	}
	return m[2], nil // The title is the second subexpression
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
  http.HandleFunc("/", redirectToFrontPageHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
