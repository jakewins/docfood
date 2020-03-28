package main

import (
	"context"
	"encoding/json"
	"feed/pkg/store"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"
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

	httpPort, found := os.LookupEnv("HTTP_PORT")
	if !found {
		httpPort = "8080"
	}

	baseCtx := context.Background()

	datadir := os.Getenv("DATA_DIR")
	if datadir == "" {
		panic("Need DATA_DIR set pls")
	}

	db, err := store.NewFileStore(path.Join(datadir, "subscriptions"))
	if err != nil {
		panic(err)
	}

	baseCtx = store.NewContext(baseCtx, db)

	if os.Getenv("HTTPS_PORT") != "" {
		startHttpsServer(router, baseCtx, os.Getenv("HTTPS_PORT"), httpPort)
	} else {
		startHttpServer(httpPort, router, baseCtx)
	}
}

func startHttpServer(port string, handler http.Handler, baseCtx context.Context) {
	server := &http.Server{Addr: ":" + port, Handler: handler}
	server.BaseContext = func(config net.Listener) context.Context { return baseCtx }

	log.Println("Running at 0.0.0.0:" + port)
	log.Fatal(server.ListenAndServe())
}

func startHttpsServer(handler http.Handler, baseCtx context.Context, httpsPort, httpPort string) {
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}
	srv.BaseContext = func(config net.Listener) context.Context { return baseCtx }

	srv.Addr = ":" + httpsPort

	go func() {
		log.Println("Running at 0.0.0.0:" + httpsPort)
		log.Fatal(srv.Serve(autocert.NewListener("www.docfood.org", "docfood.org")))
	}()

	startHttpServer(httpPort, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		target := "https://" + req.Host + req.URL.Path
		if httpsPort != "443" {
			target = "https://" + req.Host + ":" + httpsPort + req.URL.Path
		}
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		http.Redirect(w, req, target, http.StatusMovedPermanently)
	}), baseCtx)
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
