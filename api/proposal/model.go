package proposal

import (
	"log/slog"
	"time"

	"github.com/cartabinaria/polleg/models"
	"gorm.io/gorm"
)

type Proposal struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	DocumentID   string `json:"document_id"`
	DocumentPath string `json:"document_path; omitempty"`
	Start        uint32 `json:"start"`
	End          uint32 `json:"end"`

	Username string `json:"username"`
}

func dbProposalToProposal(db *gorm.DB, p *models.Proposal) Proposal {
	var user models.User
	if err := db.Find(&user, "id = ?", p.UserID).Error; err != nil {
		slog.With("user_id", p.UserID, "err", err).Error("db query failed finding user")
		user = models.User{
			Username: "unknown",
		}
	}

	return Proposal{
		ID:        p.ID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,

		DocumentID:   p.DocumentID,
		DocumentPath: p.DocumentPath,
		Start:        p.Start,
		End:          p.End,

		Username: user.Username,
	}
}

func dbProposalsToProposals(db *gorm.DB, p []models.Proposal) []Proposal {
	res := make([]Proposal, len(p))
	for i, prop := range p {
		res[i] = dbProposalToProposal(db, &prop)
	}
	return res
}
