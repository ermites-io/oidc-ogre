// +build go1.12

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func grpcServer(svc *OidcService, hostport string) {

	nfd, err := net.Listen("tcp", hostport)
	if err != nil {
		panic(err)
	}

	fmt.Printf("starting grpc service\n")
	grpcServer := grpc.NewServer()

	RegisterOidcServer(grpcServer, svc)
	err = grpcServer.Serve(nfd)
	if err != nil {
		panic(err)
	}
}

func restServer(svc *OidcService, hostport string) {
	var dialopts []grpc.DialOption

	dialopts = append(dialopts, grpc.WithInsecure())
	dialopts = append(dialopts, grpc.WithDefaultCallOptions(grpc.FailFast(true)))
	dialopts = append(dialopts, grpc.WithBlock())

	nfd, err := grpc.Dial(grpcConnect, dialopts...)
	if err != nil {
		panic(err)
	}

	grpcClient := NewOidcClient(nfd)
	if err != nil {
		panic(err)
	}

	runtime.HTTPError = svc.OidcHandleHTTPError

	grpcmux := runtime.NewServeMux(
		runtime.WithProtoErrorHandler(svc.OidcHandleHTTPError),    // handle our library specific error codes.
		runtime.WithForwardResponseOption(cookieOrRedirectMapper), // create the Location Header + state cookie for the auth
		runtime.WithMetadata(headerToMetadata),                    // get the cookie and fill it in metadatas.
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	)

	err = RegisterOidcHandlerClient(context.Background(), grpcmux, grpcClient)
	if err != nil {
		panic(err)
	}

	fmt.Printf("starting rest-gateway service\n")
	err = http.ListenAndServe(hostport, grpcmux) // serve that on 8080
	panic(err)
}

func restOnlyServer() {
}
