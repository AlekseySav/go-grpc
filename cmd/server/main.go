package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"blog/gen/blog"
	"blog/internal/handler"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. gRPC сервер
	lis, _ := net.Listen("tcp", ":50051")
	grpcSrv := grpc.NewServer()
	reflection.Register(grpcSrv)
	blog.RegisterBlogServiceServer(grpcSrv, handler.NewServer())
	go func() { fmt.Println("gRPC listening on :50051"); grpcSrv.Serve(lis) }()

	// 2. gRPC-Gateway
	mux := runtime.NewServeMux()
	blog.RegisterBlogServiceHandlerFromEndpoint(
		context.Background(), mux, "localhost:50051",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)

	// 3. HTTP роутер (API + Swagger)
	httpMux := http.NewServeMux()
	httpMux.Handle("/v1/", mux)
	httpMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))

	fmt.Println("HTTP Gateway + Swagger listening on :8080")
	http.ListenAndServe(":8080", httpMux)
}
