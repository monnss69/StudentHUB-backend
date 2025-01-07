package interfaces

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"type:varchar(50);unique;not null"`
	Email        string    `json:"email" gorm:"type:varchar(255);unique;not null"`
	PasswordHash string    `json:"password_hash" gorm:"column:password_hash;type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type AuthenticateUser struct {
	Username     string `json:"username" gorm:"type:varchar(50);not null"`
	PasswordHash string `json:"password" gorm:"column:password_hash;type:varchar(255);not null"`
}

type Category struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(100);unique;not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type Post struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title      string    `json:"title" gorm:"type:varchar(255);not null"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	AuthorID   uuid.UUID `json:"author_id" gorm:"column:author_id;type:uuid;not null;references:users(id)"`
	CategoryID uuid.UUID `json:"category_id" gorm:"column:category_id;type:uuid;not null;references:categories(id)"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
	UpdatedAt  time.Time `gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(50);unique;not null"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type PostTag struct {
	PostID uuid.UUID `gorm:"column:post_id;type:uuid;primary_key;references:posts(id)"`
	TagID  uuid.UUID `gorm:"column:tag_id;type:uuid;primary_key;references:tags(id)"`
}

type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	PostID    uuid.UUID `json:"post_id" gorm:"column:post_id;type:uuid;not null;references:posts(id)"`
	AuthorID  uuid.UUID `json:"author_id" gorm:"column:author_id;type:uuid;not null;references:users(id)"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type CommentInput struct {
	Content  string    `json:"content"`
	AuthorID uuid.UUID `json:"author_id"`
}
