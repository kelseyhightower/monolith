package foo

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/logging"
	"go.opencensus.io/plugin/ochttp"
)

type handler struct {
	client *http.Client
	logger *logging.Logger
}

// Server returns a handler that HTTP requests.
func Server(logger *logging.Logger) http.Handler {
	client := &http.Client{
		Transport: &ochttp.Transport{},
	}

	return &handler{client, logger}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := http.NewRequest("GET", "http://127.0.0.1/bar", nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	request = request.WithContext(r.Context())

	response, err := h.client.Do(request)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer response.Body.Close()

	fmt.Fprintf(w, string(data))
	fmt.Fprintf(w, "Foo Service")
}
