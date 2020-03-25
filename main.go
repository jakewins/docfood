package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"feed/pkg/store"
	"fmt"
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

	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
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

	if os.Getenv("PRODUCTION") == "true" {
		tlsCacheDir := os.Getenv("TLS_CACHE_DIR")
		if tlsCacheDir == "" {
			panic("Need TLS_CACHE_DIR set pls")
		}
		startHttpsServer(tlsCacheDir, router, port)
	} else {
		server := &http.Server{Addr: ":" + port, Handler: router}
		server.BaseContext = func(config net.Listener) context.Context { return baseCtx }

		log.Println("Running at 0.0.0.0:" + port)
		log.Fatal(server.ListenAndServe())
	}
}

func startHttpsServer(cacheDir string, handler http.Handler, port string) {
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			// Note: change to your real domain
			allowedHost := "www.docfood.org"
			if host == allowedHost {
				return nil
			}
			return fmt.Errorf("acme/autocert: only %s host is allowed", allowedHost)
		},
		Cache:      autocert.DirCache(cacheDir),
	}
	srv.Addr = ":" + port
	srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

	go func() {
		log.Println("Running at 0.0.0.0:" + port)
		log.Fatal(srv.ListenAndServeTLS("", ""))
	}()
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
