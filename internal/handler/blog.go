package handler

import (
	"context"

	pb "blog/gen/blog"
	"blog/internal/repository"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedBlogServiceServer
	repo *repository.PostRepository
}

func NewServer(repo *repository.PostRepository) *Server {
	return &Server{repo: repo}
}

func getUserID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("x-user-id"); len(v) > 0 {
			return v[0]
		}
	}
	return "anon"
}

func (s *Server) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	posts, err := s.repo.GetPosts(ctx, getUserID(ctx), req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	return &pb.GetPostsResponse{Posts: posts}, nil
}

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	userID := getUserID(ctx)
	return s.repo.CreatePost(ctx, userID, userID, "", req.Body)
}

func (s *Server) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.Post, error) {
	return s.repo.UpdatePost(ctx, req.Id, req.Body)
}

func (s *Server) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.repo.DeletePost(ctx, req.Id)
}

func (s *Server) ToggleLike(ctx context.Context, req *pb.ToggleLikeRequest) (*pb.Post, error) {
	return s.repo.ToggleLike(ctx, req.Id, getUserID(ctx))
}