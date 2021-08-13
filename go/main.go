package main

import "C"

// other imports should be seperate from the special Cgo import
import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// Embed the file content as string.
//go:embed title.txt
var title string

// Embed the entire directory.
//go:embed templates
var indexHTML embed.FS

//go:embed static
var staticFiles embed.FS

//export server
func server() {
	c := make(chan bool)

	// http.FS can be used to create a http Filesystem
	var staticFS = http.FS(staticFiles)
	fs := http.FileServer(staticFS)

	// Serve static files
	http.Handle("/static/", fs)

	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
		<-c
	}()

	http.HandleFunc("/", handler)
	http.HandleFunc("/Sayhi", HelloHandler)

}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, there from go\n")
}

func handler(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	// Note the call to ParseFS instead of Parse
	t, err := template.ParseFS(indexHTML, "templates/index.html.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "text/html")

	// respond with the output of template execution
	t.Execute(w, struct {
		Title    string
		Response string
	}{Title: title, Response: path})

}

func main() {}
