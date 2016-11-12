package main

import (
	"github.com/smacker/newrelic-context"
	"log"
	"net/http"
)

func indexHandlerFunc(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("I'm an index page!"))

	client := &http.Client{Timeout: 10}
	nrcontext.WrapHTTPClient(req.Context(), client)
	_, err := client.Get("http://google.com")
	if err != nil {
		rw.Write([]byte("Can't fetch google :("))
		return
	}
	rw.Write([]byte("Google fetched successfully!"))
}

func main() {
	var handler http.Handler

	handler = http.HandlerFunc(indexHandlerFunc)
	nrmiddleware, err := nrcontext.NewMiddleware("test-app", "my-license-key")
	if err != nil {
		log.Print("Can't create newrelic middleware: ", err)
	} else {
		handler = nrmiddleware.Handler(handler)
	}

	http.ListenAndServe(":8000", handler)
}
