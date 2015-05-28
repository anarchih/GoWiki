// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
    "flag"
    "html/template"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "regexp"
    "fmt"
    "strings"
    "os"
)

var (
    addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)
type Page interface {
    save() error
    del() error
}

type SimplePage struct {
    Title string
    Body  []byte
}

type IndexPage struct {
    Title string
    Body  [][]byte
}

func (p *SimplePage) save() error {
    filename := "data/" + p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func (p *SimplePage) del() error {
    filename := "data/" + p.Title + ".txt"
    return os.Remove(filename)
}

func (p *IndexPage) save() error {
    filename := "tester"
    return ioutil.WriteFile(filename, p.Body[0], 0600)
}

func (p *IndexPage) del() error {
    filename := "data/" + p.Title + ".txt"
    return os.Remove(filename)
}

func loadPage(title string) (Page, error) {
    filename := "data/" + title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &SimplePage{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    if title != ""{
        p, err := loadPage(title)
        if err != nil {
            http.Redirect(w, r, "/edit/"+title, http.StatusFound)
            return
        }
        renderTemplate(w, "view", p)
    } else {
        indexHandler(w, r, title)
    }
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if title == "" {
        http.Redirect(w, r, "/", http.StatusFound)
    }
    if err != nil {
        p = &SimplePage{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &SimplePage{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/"+title, http.StatusFound)
}

func delHandler(w http.ResponseWriter, r *http.Request, title string) {
    p := &SimplePage{Title: title, Body: []byte("")}
    _ = p.del()
    http.Redirect(w, r, "/", http.StatusFound)
}

func indexHandler(w http.ResponseWriter, r *http.Request, title string) {
    info, _ := ioutil.ReadDir("data/")
    var byteBody [][]byte
    for i := 0; i < len(info); i++ {
        tmp := info[i].Name() + " "
        tmp = strings.Replace(tmp, ".txt ", "", -1)
        byteBody = append(byteBody, [][]byte{[]byte(tmp)}...)
    }

    p := &IndexPage{Title:"index", Body:byteBody}
    renderTemplate(w, "index", p)
}

var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html", "templates/index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

var validPath = regexp.MustCompile("(del|edit|save|view|index| *)/([a-zA-Z0-9]*)$")

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

func main() {
    flag.Parse()
    http.HandleFunc("/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/del/", makeHandler(delHandler))
    http.HandleFunc("/index/", makeHandler(indexHandler))
    if *addr {
        l, err := net.Listen("tcp", "127.0.0.1:0")
        if err != nil {
            log.Fatal(err)
        }
        err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
        if err != nil {
            log.Fatal(err)
        }
        s := &http.Server{}
        s.Serve(l)
        return
    }

    fmt.Println("Running")
    http.ListenAndServe(":8080", nil)
}
