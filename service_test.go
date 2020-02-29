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
