// +build go1.12

package main

//go:generate protoc oidc.proto -I. -I$HOME/tools/pb3/include -I/home/rival/dev/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.12.1/third_party/googleapis --go_out=plugins=grpc:.
//go:generate protoc oidc.proto -I. -I$HOME/tools/pb3/include -I/home/rival/dev/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.12.1/third_party/googleapis --grpc-gateway_out=logtostderr=true:.

import (
	"flag"
	"fmt"

	"github.com/ermites-io/oidc"
)

const (
	grpcConnect = "127.0.0.1:8888"
	restConnect = "127.0.0.1:8000"
	oauthConfig = "oauth_conf.json"
)

//
//
// we will listen on 2 ports
// 8888: GRPC service
// 8000: REST gateway

func main() {
	fmt.Printf("Openid Grpc Rp Example service\n")
	fmt.Printf("this is an implementation example using GRPC & GRPC-gateway of an OpenID authentication party\n")

	config := flag.String("config", "openid-configuration", "openid configuration")
	name := flag.String("name", "", "provider name")
	clientIdFlag := flag.String("id", "", "client id")
	clientSecretFlag := flag.String("secret", "", "client secret")
	clientRedirectFlag := flag.String("redirect", "", "redirect url configured")

	flag.Parse()
	//argv := flag.Args()

	idp, err := oidc.NewProvider(*name, *config)
	if err != nil {
		panic(err)
	}

	// setup the new provider..
	err = idp.SetAuth(*clientIdFlag, *clientSecretFlag, *clientRedirectFlag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Provider: %v\n", idp.GetName())

	// add providers
	svc := NewOidcService(idp)
	svc.SetFailUrl("https://login.ermite.io/login-failed")
	svc.SetOkUrl("https://login.ermite.io/login-ok")

	//
	// start grpc
	go grpcServer(svc, grpcConnect)

	// then rest
	restServer(svc, restConnect)
}
