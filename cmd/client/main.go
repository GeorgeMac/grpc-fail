package main

import (
	"log"
	"os"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
		grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer)),
	}

	conn, err := grpc.Dial(os.Getenv("SERVER_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}
}
