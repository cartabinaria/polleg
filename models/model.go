package models

import (
	"time"

	"gorm.io/gorm"
)

type Answer struct {
	// taken from from gorm.Model, so we can json strigify properly
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Question uint  `json:"question" gorm:"foreignKey:Question;references:ID"`
	Parent   *uint `json:"parent"`

	UserId    uint        `json:"-"`
	Content   string      `json:"content"`
	Upvotes   uint32      `json:"upvotes" gorm:"->"`
	Downvotes uint32      `json:"downvotes" gorm:"->"`
	Replies   []Answer    `json:"replies" gorm:"foreignKey:Parent;references:ID"`
	Votes     []Vote      `json:"-" gorm:"foreignKey:Answer;references:ID"`
	Anonymous bool        `json:"anonymous"`
	State     AnswerState `json:"state"`
}

type AnswerState uint8

const (
	AnswerStateVisible AnswerState = iota
	AnswerStateDeletedByUser
	AnswerStateDeletedByAdmin
)

type Question struct {
	// taken from from gorm.Model, so we can json strigify properly
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Document string   `json:"document"`
	Start    uint32   `json:"start"`
	End      uint32   `json:"end"`
	Answers  []Answer `json:"answers" gorm:"foreignKey:Question;references:ID;constraint:OnDelete:CASCADE;"`
}

type Vote struct {
	Answer uint `json:"answer" gorm:"primaryKey"`
	UserId uint `json:"-" gorm:"primaryKey"`
	Vote   int8 `json:"vote"`

	// taken from from gorm.Model
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID       uint   `json:"-" gorm:"primarykey"`
	Username string `json:"username"`
	Alias    string `json:"alias"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type PostAnswerRequest struct {
	Question  uint   `json:"question"`
	Parent    *uint  `json:"parent"`
	Content   string `json:"content"`
	Anonymous bool   `json:"anonymous"`
}

type AnswerResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Question uint  `json:"question"`
	Parent   *uint `json:"parent"`

	User          string           `json:"user"`
	UserAvatarURL string           `json:"user_avatar_url"`
	Content       string           `json:"content"`
	Upvotes       uint32           `json:"upvotes"`
	Downvotes     uint32           `json:"downvotes"`
	Replies       []AnswerResponse `json:"replies"`
}

type VoteValue int8

type PostVoteRequest struct {
	Vote VoteValue `json:"vote"`
}

type VoteResponse struct {
	Answer uint   `json:"answer"`
	User   string `json:"user"`
	Vote   int8   `json:"vote"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type QuestionResponse struct {
	ID        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`

	Document string           `json:"document"`
	Start    uint32           `json:"start"`
	End      uint32           `json:"end"`
	Answers  []AnswerResponse `json:"answers"`
}
