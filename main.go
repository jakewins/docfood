package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var indexTemplate = mustLoadTemplate("index", "templates/index.html")
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := indexTemplate.Execute(w, nil); err != nil {
		log.Printf("Failed to render index: %s\n", err)
		w.WriteHeader(503)
	}
}

type subscribePayload struct {
	Email               string
	PaymentMethod       string
	AllRestaurants      bool
	SpecificRestaurants []string
	Subscription        struct {
		SubType string
		Amount  string
	}
}

func Subscribe(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	payload := &subscribePayload{}
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		log.Printf("Failed to read payload: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("subscription received: %v\n", payload)

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte("{\"result\": \"ok\"}")); err != nil {
		log.Printf("Failed to write subscribe response for %v: %s\n", payload, err)
		return
	}
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/v1/subscribe", Subscribe)

	router.NotFound = http.FileServer(http.Dir("./static"))

	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}

	log.Println("Running at 0.0.0.0:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func mustLoadTemplate(name, path string) *template.Template {
	tmpl, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	t, err := template.New(name).Parse(string(tmpl))
	if err != nil {
		panic(err)
	}
	return t
}
