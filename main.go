package main

import (
	"context"
	"encoding/json"
	"feed/pkg/store"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"log"
	"net"
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

	db, ok := store.FromContext(r.Context())
	if ! ok {
		log.Printf("FATAL: No store configured in context\n")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	err = db.CreateSubscription(store.Subscription{
		Email:               payload.Email,
		AllRestaurants:      payload.AllRestaurants,
		SpecificRestaurants: payload.SpecificRestaurants,
		SubscriptionType:    payload.Subscription.SubType,
		Amount:              payload.Subscription.Amount,
		PaymentMethod:       payload.PaymentMethod,
	})
	if err != nil {
		log.Printf("FATAL: Failed to create subscription: %s\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

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


	baseCtx := context.Background()

	var db store.Store
	if os.Getenv("PRODUCTION") != "" {
		db = store.NewFirestore()
	} else {
		db = store.NewMemStore()
	}

	baseCtx = store.NewContext(baseCtx, db)

	server := &http.Server{Addr: ":"+port, Handler: router}
	server.BaseContext = func(config net.Listener) context.Context { return baseCtx }

	log.Println("Running at 0.0.0.0:" + port)

	log.Fatal(server.ListenAndServe())
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
