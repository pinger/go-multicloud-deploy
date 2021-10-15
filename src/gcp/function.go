// Package p contains an HTTP Cloud Function.
package p

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pinger/go-multicloud-deploy/src/function/v2"
)

// HelloWorld prints the JSON encoded "message" field in the body
// of the request or "Hello, World!" if there isn't one.
func EndPoint01(w http.ResponseWriter, r *http.Request) {
	var d function.Event

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		switch err {
		case io.EOF:
			fmt.Fprint(w, "Hello World!")
			return
		default:
			log.Printf("json.NewDecoder: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	if d.Message == "" {
		fmt.Fprint(w, "Hello World!")
		return
	}

	// all good. write our message.
	j, _ := json.Marshal(d)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(string(j)))
}
