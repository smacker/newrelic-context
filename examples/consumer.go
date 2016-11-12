package main

import (
	"context"
	"fmt"
	"github.com/smacker/newrelic-context"
	"log"
	"net/http"
)

func Consume(ctx context.Context, query string) {
	client := &http.Client{Timeout: 10}
	nrcontext.WrapHTTPClient(ctx, client)
	_, err := client.Get(fmt.Sprintf("https://www.google.com.vn/?q=%v", query))
	if err != nil {
		log.Println("Can't fetch google :(")
		return
	}
	log.Println("Google fetched successfully!")
}
