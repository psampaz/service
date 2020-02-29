package service

import (
	"context"
)

// Request is the request that the service will serve.
type Request struct {
	// Sample field for the sake of the example. Could be on or more fields of any type.
	Data string
}

// Response is the actual reponse of the service in absence of error (happy path)
type Response struct {
	// Sample field for the sake of the example. Could be on or more fields of any type.
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
