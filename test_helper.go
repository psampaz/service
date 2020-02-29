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
