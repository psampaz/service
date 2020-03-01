![Build Status](https://github.com/psampaz/service/workflows/build/badge.svg)
[![GoDoc](https://godoc.org/github.com/psampaz/service?status.svg)](https://godoc.org/github.com/psampaz/service)
[![Go Report Card](https://goreportcard.com/badge/github.com/psampaz/service)](https://goreportcard.com/report/github.com/psampaz/service)

A fully testable and commented Go service prototype built for educational purposes.

## Service
```go
package service

import (
	"context"
)

// Request is the request that the service will serve.
type Request struct {
	// Sample field for the sake of the example. Could be one or more fields of any type.
	Data string
}

// Response is the actual reponse of the service in absence of error (happy path)
type Response struct {
	// Sample field for the sake of the example. Could be one or more fields of any type.
	Data string
}

// Service is a struct representing the actual service. For the sake of the example it has only one field
// which simulates the work that needs to be completed.
type Service struct {
	// func representing the actual work that needs to be done in order to calculate the response.
	// Could be an external HTTP call, db interaction, data processing or whatever else.
	work func() (Response, error)
}

// NewService is a factory function/constructor for the Service
func NewService(work func() (Response, error)) *Service {
	return &Service{
		work: work,
	}
}

// Serve is the method of the Service that handles the request.
// It responds back with a Response on the happy  path or an error in case of failure
func (s *Service) Serve(ctx context.Context, req Request) (Response, error) {
	// Use buffered channel to avoid goroutine leak in case the context gets cancelled
	// Read this excellent article for more details:
	// https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html
	resCh := make(chan Response, 1)
	errCh := make(chan error, 1)

	go func() {
		// Do the work.
		// In case of an error send the error in the errCh and return
		resp, err := s.work()
		if err != nil {
			errCh <- err
			return
		}

		// In case of happy path send the actual response in the resCh channel
		resCh <- resp
	}()
	// Select will block until there is a errCh or resCh receives a message or the context is cancelled
	// due to a timeout, deadline on direct cancellation (using the cancel function)
	select {
	case err := <-errCh:
		return Response{}, err
	case res := <-resCh:
		return res, nil
	case <-ctx.Done():
		return Response{}, ctx.Err()
	}
}
```

## Service tests
```go
package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

// Test case for the happy path. The service served the request in time without errors.
func TestService_Serve_Success(t *testing.T) {
	srv := NewService(func() (Response, error) {
		time.Sleep(500 * time.Millisecond)
		return Response{Data: "success"}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	response, err := srv.Serve(ctx, Request{})

	if err != nil {
		t.Errorf("Serve() should not return an error, go %v", err)
	}

	wantResp := Response{"success"}
	if !reflect.DeepEqual(response, wantResp) {
		t.Errorf("Serve() got response %v, wanted %v", response, wantResp)
	}
}

// Test case for service failure. Service failed to serve the request before reaching the context timeout.
func TestService_Serve_Error(t *testing.T) {
	wantErr := errors.New("error")
	srv := NewService(func() (Response, error) {
		time.Sleep(500 * time.Millisecond)
		return Response{}, wantErr
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	response, err := srv.Serve(ctx, Request{})

	if err == nil {
		t.Errorf("Serve() got err %v, wanted %v", err, wantErr)
	}

	wantResp := Response{}
	if !reflect.DeepEqual(response, wantResp) {
		t.Errorf("Serve() got response %v, wanted %v", response, wantResp)
	}
}

// Test case for service timeout. Context timed out before the service finished serving the request.
func TestService_Serve_Timeout(t *testing.T) {

	srv := NewService(func() (Response, error) {
		time.Sleep(2000 * time.Millisecond)
		return Response{Data: "success"}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	response, err := srv.Serve(ctx, Request{})

	if err == nil {
		t.Errorf("Serve() got err %v, wanted %v", err, context.DeadlineExceeded)
	}

	wantResp := Response{}
	if !reflect.DeepEqual(response, wantResp) {
		t.Errorf("Serve() got response %v, wanted %v", response, wantResp)
	}
}
```

# Test Helper

If you have a dependency on a Service you can use the following test helper to make you life easier.
You can control the response of the error of the service, the timeout of the context and make assertions on the request 
served and the context cancellation and error.
```go
package service

import (
	"context"
	"errors"
	"time"
)

// Server is an interface to use in your code in order to be able to switch implementations
// between application and testing code
type Server interface {
	Serve(ctx context.Context, req Request) (Response, error)
}

// TestService is an implementation of the Server interface for testing purposes
type TestService struct {
	// The response that should be returned
	Res Response
	// DelayReponse is the time to delay the response of the test service.
	// Should be used when testing with cancellable context
	DelayReponse time.Duration
	// Err is the error that should be returned
	Err error
	// Recorder stores informations about the Serve execution
	Recorder struct {
		// Request is the actual request that was served
		Request Request
		// CtxCancelled is a flag showing if the context was cancelled or not
		CtxCancelled bool
		// CtxCancelled is a flag showing if the context exceeded a deadline
		CtxDeadlineExceeded bool
		// CtxErr is the error returned in case of context cancellation.
		CtxErr error
	}
}

// Serve serves and records the request and context cancellation and error, and replys back with
// a predefined response or error
func (t *TestService) Serve(ctx context.Context, req Request) (Response, error) {
	// record the request param
	t.Recorder.Request = req

	// create a channel to signal that the actual work was finished
	done := make(chan bool, 1)
	go func() {
		time.Sleep(t.DelayReponse)
		done <- true
	}()

	select {
	case <-ctx.Done():
		t.Recorder.CtxErr = ctx.Err()
		if errors.Is(ctx.Err(), context.Canceled) {
			t.Recorder.CtxCancelled = true
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Recorder.CtxDeadlineExceeded = true
		}
		return Response{}, ctx.Err()
	case <-done:
		return t.Res, t.Err
	}
}
``` 
# Examples

## Example of Service use
```go
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
```

## Example of TestService use 

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/psampaz/service"
)

func main() {
	// Create a test service that return a response immediately
	th1 := service.TestService{
		Res:          service.Response{Data: "response data"},
		DelayReponse: 0,
		Err:          nil,
	}
	// create a context without any timeout
	th1.Serve(context.Background(), service.Request{Data: "request data"})

	fmt.Printf("%+v", th1.Recorder)
	// {
	//  Request:{Data:request data}
	//  CtxCancelled:false
	//  CtxDeadlineExceeded:false
	//  CtxErr:<nil>
	// }

	// Create a test service that will delay the response for 1 second
	th2 := service.TestService{
		Res:          service.Response{Data: "response data"},
		DelayReponse: time.Second,
		Err:          nil,
	}

	// create a context that will timeout in 1 millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	th2.Serve(ctx, service.Request{Data: "request data"})

	fmt.Printf("%+v", th2.Recorder)
	// {
	//  Request:{Data:request data}
	//  CtxCancelled:false
	//  CtxDeadlineExceeded:true
	//  CtxErr:context deadline exceeded
	// }
}
```
