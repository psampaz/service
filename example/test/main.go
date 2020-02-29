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
