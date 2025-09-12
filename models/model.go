package models

import (
	"time"

	"gorm.io/gorm"
)

type Answer struct {
	// taken from from gorm.Model, so we can json strigify properly
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Question uint `gorm:"foreignKey:Question;references:ID"`
	Parent   *uint

	UserId    uint
	Upvotes   uint32   `gorm:"->"`
	Downvotes uint32   `gorm:"->"`
	Replies   []Answer `gorm:"foreignKey:Parent;references:ID"`
	Votes     []Vote   `gorm:"foreignKey:AnswerID;references:ID"`
	Anonymous bool
	State     AnswerState
}

type AnswerVersion struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	AnswerID uint `gorm:"foreignKey:Answer;references:ID;index;not null"`
	Content  string
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
	AnswerID uint `gorm:"primaryKey"`
	UserId   uint `gorm:"primaryKey"`
	Vote     int8

	// taken from from gorm.Model
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID       uint `gorm:"primarykey"`
	Username string
	Alias    string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Banned   bool `gorm:"default:false"`
	BannedAt *time.Time

	Questions []Question `gorm:"foreignKey:UserID;references:ID"`
	Proposals []Proposal `gorm:"foreignKey:UserID;references:ID"`
	Reports   []Report   `gorm:"foreignKey:UserID;references:ID"`
}

type PostAnswerRequest struct {
	Question  uint
	Parent    *uint
	Content   string
	Anonymous bool
}

type UpdateAnswerRequest struct {
	Content string
}

type Image struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

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

type Report struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	AnswerID uint `gorm:"index; not null;"`
	Cause    string
	UserID   uint `gorm:"index; not null;"`
}
