// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package models

import (
	"time"
)

type Comment struct {
	ID        int        `json:"id"`
	Payload   string     `json:"payload"`
	PostID    int        `json:"postID"`
	Author    *User      `json:"author"`
	ReplyTo   *int       `json:"replyTo,omitempty"`
	Replies   []*Comment `json:"replies,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type Mutation struct {
}

type NewComment struct {
	Payload  string `json:"payload"`
	PostID   int    `json:"postID"`
	AuthorID int    `json:"authorID"`
	ReplyTo  *int   `json:"replyTo,omitempty"`
}

type NewPost struct {
	Title             string `json:"title"`
	Payload           string `json:"payload"`
	AuthorID          int    `json:"authorID"`
	IsCommentsAllowed bool   `json:"IsCommentsAllowed"`
}

type Post struct {
	ID                int        `json:"id"`
	Title             string     `json:"title"`
	Payload           string     `json:"payload"`
	Author            *User      `json:"author"`
	IsCommentsAllowed bool       `json:"isCommentsAllowed"`
	Comments          []*Comment `json:"comments,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type Query struct {
}

type Subscription struct {
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}
