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
	Upvotes   uint32      `json:"upvotes" gorm:"->"`
	Downvotes uint32      `json:"downvotes" gorm:"->"`
	Replies   []Answer    `json:"replies" gorm:"foreignKey:Parent;references:ID"`
	Votes     []Vote      `json:"-" gorm:"foreignKey:AnswerID;references:ID"`
	Anonymous bool        `json:"anonymous"`
	State     AnswerState `json:"state"`
}

type AnswerVersion struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`

	AnswerID uint   `gorm:"foreignKey:Answer;references:ID;index;not null"`
	Content  string `json:"content"`
}

type AnswerState uint8

const (
	AnswerStateVisible AnswerState = iota
	AnswerStateDeletedByUser
	AnswerStateDeletedByAdmin
)

type Question struct {
	// taken from from gorm.Model, so we can json strigify properly
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Document string
	Start    uint32
	End      uint32
	Answers  []Answer `gorm:"foreignKey:Question;references:ID"`

	UserID uint `gorm:"index; not null;"`
}

func (q *Question) AfterDelete(tx *gorm.DB) (err error) {
	tx.Model(&Answer{}).Where("question = ?", q.ID).Update("state", AnswerStateDeletedByAdmin)
	return
}

type Vote struct {
	AnswerID uint `json:"answer" gorm:"primaryKey"`
	UserId   uint `json:"-" gorm:"primaryKey"`
	Vote     int8 `json:"vote"`

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

	Questions []Question `json:"-" gorm:"foreignKey:UserID;references:ID"`
	Proposals []Proposal `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

type PostAnswerRequest struct {
	Question  uint   `json:"question"`
	Parent    *uint  `json:"parent"`
	Content   string `json:"content"`
	Anonymous bool   `json:"anonymous"`
}

type UpdateAnswerRequest struct {
	Content string `json:"content"`
}

type Image struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `json:"-"`

	UserID uint `gorm:"index; not null; foreignKey:User; references:ID"`
	Size   uint `gorm:"not null"`
}

type Proposal struct {
	// taken from from gorm.Model, so we can json strigify properly
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	DocumentID   string
	DocumentPath string
	Start        uint32
	End          uint32

	UserID uint `gorm:"index; not null;"`
}
