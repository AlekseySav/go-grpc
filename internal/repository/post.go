package repository

import (
	"context"
	"fmt"
	"time"

	"blog/internal/db"
	pb "blog/gen/blog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(database *gorm.DB) *PostRepository {
	return &PostRepository{db: database}
}

type postRow struct {
	db.Post
	LikesCount    int32
	IsLikedByUser bool
}

func (r *PostRepository) GetPosts(ctx context.Context, userID string, limit, offset int32) ([]*pb.Post, error) {
	if limit == 0 {
		limit = 20
	}
	var rows []postRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.*,
		       COUNT(l.user_id)::int AS likes_count,
		       EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ?) AS is_liked_by_user
		FROM posts p
		LEFT JOIN likes l ON l.post_id = p.id
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	posts := make([]*pb.Post, len(rows))
	for i, row := range rows {
		posts[i] = toProto(row.Post, row.LikesCount, row.IsLikedByUser)
	}
	return posts, nil
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
	var count int64
	r.db.Model(&db.Like{}).Where("post_id = ?", id).Count(&count)
	return toProto(post, int32(count), false), nil
}

func (r *PostRepository) DeletePost(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Where("post_id = ?", id).Delete(&db.Like{})
		return tx.Where("id = ?", id).Delete(&db.Post{}).Error
	})
}

func (r *PostRepository) ToggleLike(ctx context.Context, postID, userID string) (*pb.Post, error) {
	var existing db.Like
	err := r.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		r.db.WithContext(ctx).Create(&db.Like{PostID: postID, UserID: userID})
	} else if err == nil {
		r.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).Delete(&db.Like{})
	} else {
		return nil, err
	}

	var rows []postRow
	err = r.db.WithContext(ctx).Raw(`
		SELECT p.*,
		       COUNT(l.user_id)::int AS likes_count,
		       EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ?) AS is_liked_by_user
		FROM posts p
		LEFT JOIN likes l ON l.post_id = p.id
		WHERE p.id = ?
		GROUP BY p.id
	`, userID, postID).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("post not found")
	}
	return toProto(rows[0].Post, rows[0].LikesCount, rows[0].IsLikedByUser), nil
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