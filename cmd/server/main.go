package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"blog/gen/blog"
	"blog/internal/cache"
	"blog/internal/db"
	"blog/internal/handler"
	"blog/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=blog port=5432 sslmode=disable"
	}
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	database, err := db.Connect(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis connect: %v\n", err)
		os.Exit(1)
	}

	repo := repository.NewPostRepository(database, cache.NewLikeCache(rdb), cache.NewPostsCache(rdb))

	lis, _ := net.Listen("tcp", ":50051")
	grpcSrv := grpc.NewServer()
	reflection.Register(grpcSrv)
	blog.RegisterBlogServiceServer(grpcSrv, handler.NewServer(repo))
	go func() { fmt.Println("gRPC listening on :50051"); grpcSrv.Serve(lis) }()

	mux := runtime.NewServeMux()
	blog.RegisterBlogServiceHandlerFromEndpoint(
		context.Background(), mux, "localhost:50051",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)

	httpMux := http.NewServeMux()
	httpMux.Handle("/v1/", mux)
	httpMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))

	fmt.Println("HTTP Gateway + Swagger listening on :8080")
	http.ListenAndServe(":8080", httpMux)
}