package main

import (
	"encoding/json"
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

type subscribePayload struct {
	Email string
	PaymentMethod string
	AllRestaurants bool
	SpecificRestaurants []string
	Subscription struct {
		SubType string
		Amount string
	}
}

func Subscribe(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	payload := &subscribePayload{}
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		panic(err)
	}

	fmt.Printf("subscription received: %v\n", payload)

	w.WriteHeader(201)
	if _, err = w.Write([]byte("{\"result\": \"ok\"}")); err != nil {
		panic(err)
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
	log.Fatal(http.ListenAndServe(":" + port, router))
}