package bar

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/logging"
	"go.opencensus.io/trace"
)

type handler struct {
	logger *logging.Logger
}

// Server returns a handler that HTTP requests.
func Server(logger *logging.Logger) http.Handler {
	return &handler{logger}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := trace.StartSpan(r.Context(), "foo")
	defer span.End()
	fmt.Fprintf(w, "Bar Service")
}
