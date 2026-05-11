package repository

import (
	"context"
	"fmt"
	"time"

	"blog/internal/cache"
	"blog/internal/db"
	pb "blog/gen/blog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepository struct {
	db    *gorm.DB
	likes *cache.LikeCache
	posts *cache.PostsCache
}

func NewPostRepository(database *gorm.DB, likes *cache.LikeCache, posts *cache.PostsCache) *PostRepository {
	return &PostRepository{db: database, likes: likes, posts: posts}
}

func (r *PostRepository) GetPosts(ctx context.Context, userID string, limit, offset int32) ([]*pb.Post, error) {
	if limit == 0 {
		limit = 20
	}

	dbPosts, hit, err := r.posts.Get(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	if !hit {
		if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(int(limit)).Offset(int(offset)).Find(&dbPosts).Error; err != nil {
			return nil, err
		}
		_ = r.posts.Set(ctx, limit, offset, dbPosts)
	}

	ids := make([]string, len(dbPosts))
	for i, p := range dbPosts {
		ids[i] = p.ID
	}

	stats, err := r.likes.GetStats(ctx, userID, ids)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.Post, len(dbPosts))
	for i, p := range dbPosts {
		s := stats[p.ID]
		result[i] = toProto(p, int32(s.Count), s.Liked)
	}
	return result, nil
}

func (r *PostRepository) CreatePost(ctx context.Context, authorID, nickname, photoURL, body string) (*pb.Post, error) {
	post := db.Post{
		ID:             uuid.New().String(),
		AuthorID:       authorID,
		AuthorNickname: nickname,
		AuthorPhotoURL: photoURL,
		Body:           body,
		CreatedAt:      time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(&post).Error; err != nil {
		return nil, err
	}
	_ = r.posts.Invalidate(ctx)
	return toProto(post, 0, false), nil
}

func (r *PostRepository) UpdatePost(ctx context.Context, id, body string) (*pb.Post, error) {
	var post db.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}
	post.Body = body
	if err := r.db.WithContext(ctx).Save(&post).Error; err != nil {
		return nil, err
	}
	_ = r.posts.Invalidate(ctx)
	stats, err := r.likes.GetStats(ctx, "", []string{id})
	if err != nil {
		return nil, err
	}
	s := stats[id]
	return toProto(post, int32(s.Count), false), nil
}

func (r *PostRepository) DeletePost(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&db.Post{}).Error; err != nil {
		return err
	}
	_ = r.posts.Invalidate(ctx)
	return r.likes.Delete(ctx, id)
}

func (r *PostRepository) ToggleLike(ctx context.Context, postID, userID string) (*pb.Post, error) {
	var post db.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", postID).Error; err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}
	if _, err := r.likes.Toggle(ctx, postID, userID); err != nil {
		return nil, err
	}
	stats, err := r.likes.GetStats(ctx, userID, []string{postID})
	if err != nil {
		return nil, err
	}
	s := stats[postID]
	return toProto(post, int32(s.Count), s.Liked), nil
}

func toProto(p db.Post, likesCount int32, isLiked bool) *pb.Post {
	return &pb.Post{
		Id: p.ID,
		Author: &pb.Author{
			Id:       p.AuthorID,
			Nickname: p.AuthorNickname,
			PhotoUrl: p.AuthorPhotoURL,
		},
		Body:          p.Body,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		LikesCount:    likesCount,
		IsLikedByUser: isLiked,
	}
}