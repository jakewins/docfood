package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, err := ioutil.ReadFile("templates/index.html")
	if err != nil {
		panic(err)
	}
	t, err := template.New("index").Parse(string(tmpl))
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	router.NotFound = http.FileServer(http.Dir("./static"))

	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}

	log.Println("Running at 0.0.0.0:" + port)
	log.Fatal(http.ListenAndServe(":" + port, router))
}