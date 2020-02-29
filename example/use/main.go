package main

import (
	"context"
	"fmt"
	"time"

	"github.com/psampaz/service"
)

func main() {
	// Define a function to simulate work that needs 500 milliseconds to be completed
	work := func() (service.Response, error) {
		time.Sleep(500 * time.Millisecond)
		return service.Response{Data: "srv1 response"}, nil
	}

	srv := service.NewService(work)

	// Create a context with timeout of 1000 milliseconds.
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	request := service.Request{Data: "service 1 request"}

	// Serve the request
	response, err := srv.Serve(ctx, request)

	fmt.Printf("Respone %+v\n", response) // Respone {Data:srv1 response}
	fmt.Printf("Error %+v\n", err)        // Error <nil>
}
