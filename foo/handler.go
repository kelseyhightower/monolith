package foo

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get("http://127.0.0.1/bar")
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
