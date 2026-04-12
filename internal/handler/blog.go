package handler

import (
	"context"
	"time"

	pb "blog/gen/blog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedBlogServiceServer
}

func NewServer() *Server { return &Server{} }

func getUserID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("x-user-id"); len(v) > 0 {
			return v[0]
		}
	}
	return "anon"
}

func (s *Server) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	return &pb.GetPostsResponse{Posts: []*pb.Post{
		{Id: "1", Author: &pb.Author{Id: "a1", Nickname: "alice", PhotoUrl: "https://i.pravatar.cc/150"}, Body: "Hello!", CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339), LikesCount: 10},
		{Id: "2", Author: &pb.Author{Id: "a2", Nickname: "bob", PhotoUrl: "https://i.pravatar.cc/151"}, Body: "Go rocks", CreatedAt: time.Now().Format(time.RFC3339), LikesCount: 5, IsLikedByUser: true},
	}}, nil
}

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	return &pb.Post{Id: "new_1", Author: &pb.Author{Id: getUserID(ctx), Nickname: "user", PhotoUrl: "https://i.pravatar.cc/152"}, Body: req.Body, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}

func (s *Server) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.Post, error) {
	return &pb.Post{Id: req.Id, Author: &pb.Author{Id: getUserID(ctx), Nickname: "user", PhotoUrl: "https://i.pravatar.cc/152"}, Body: req.Body, CreatedAt: time.Now().Format(time.RFC3339), LikesCount: 1}, nil
}

func (s *Server) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) ToggleLike(ctx context.Context, req *pb.ToggleLikeRequest) (*pb.Post, error) {
	return &pb.Post{Id: req.Id, LikesCount: 6, IsLikedByUser: true}, nil
}
