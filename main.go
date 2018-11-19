package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/logging"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/kelseyhightower/gcscache"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/acme/autocert"

	"github.com/kelseyhightower/monolith/foo"
)

func main() {
	log.Println("Starting the monolith...")

	projectId := os.Getenv("GOOGLE_CLOUD_PROJECT")

	ctx := context.Background()
	loggingClient, err := logging.NewClient(ctx, projectId)
	if err != nil {
		log.Fatal(err)
	}
	defer loggingClient.Close()

	logger := loggingClient.Logger("monolith")

	stackdriverExporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: projectId})
	if err != nil {
		message := fmt.Sprintf("Failed to create Stackdriver trace exporter: %s", err)
		logger.Log(logging.Entry{
			Payload:  message,
			Severity: logging.Error,
		})
		log.Fatal(message)
	}

	trace.RegisterExporter(stackdriverExporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	cache, err := gcscache.New("hightowerlabs")
	if err != nil {
		log.Fatal(err)
	}

	m := autocert.Manager{
		Cache:      cache,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("api.hightowerlabs.com"),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Log(logging.Entry{
			HTTPRequest: &logging.HTTPRequest{Request: r},
			Payload:     "Handling HTTP request",
			Severity:    logging.Info,
		})
		fmt.Fprintf(w, "Hello World!")
	})

	mux.HandleFunc("/foo", foo.Handler)

	localServer := &http.Server{
		Addr: "127.0.0.1:80",
		Handler: &ochttp.Handler{
			Handler: mux,
		},
	}

	server := &http.Server{
		Addr: "0.0.0.0:443",
		Handler: &ochttp.Handler{
			Handler: mux,
		},
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}

	go func() {
		err := localServer.ListenAndServe()
		if err == http.ErrServerClosed {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err := server.ListenAndServeTLS("", "")
		if err == http.ErrServerClosed {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan

	localServer.Shutdown(context.Background())
	server.Shutdown(context.Background())

	log.Println("Stopping the monolith...")
}
