package proposal

import (
	"log/slog"
	"time"

	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"gorm.io/gorm"
)

type Proposal struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	DocumentID   string `json:"document"`
	DocumentPath string `json:"document_path,omitempty"`
	Start        uint32 `json:"start"`
	End          uint32 `json:"end"`

	User          string `json:"username"`
	UserAvatarURL string `json:"user_avatar_url"`
}

func dbProposalToProposal(db *gorm.DB, p *models.Proposal) Proposal {
	user, err := util.GetUserByID(db, p.UserID)
	if err != nil {
		slog.With("proposal", p, "err", err).Error("error while getting the user for a proposal")
		user = &models.User{
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

		User:          user.Username,
		UserAvatarURL: util.GetPublicAvatarURL(user.ID),
	}
}

func dbProposalsToProposals(db *gorm.DB, p []models.Proposal) []Proposal {
	res := make([]Proposal, len(p))
	for i, prop := range p {
		res[i] = dbProposalToProposal(db, &prop)
	}
	return res
}
