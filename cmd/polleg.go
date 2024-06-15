package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kataras/muxie"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/slog"

	"github.com/csunibo/auth/pkg/middleware"
	"github.com/csunibo/polleg/api"
	"github.com/csunibo/polleg/api/proposal"
	"github.com/csunibo/polleg/util"
)

type Config struct {
	Listen     string   `toml:"listen"`
	ClientURLs []string `toml:"client_urls"`

	DbURI   string `toml:"db_uri" required:"true"`
	AuthURI string `toml:"auth_uri" required:"true"`
}

var (
	// Default config values
	config = Config{
		Listen:  "0.0.0.0:3001",
		AuthURI: "http://localhost:3000",
	}
)

// @title			Polleg API
// @version		1.0
// @description	This is the backend API for Polleg that allows unibo students to answer exam exercises directly on the csunibo website
// @contact.name	Gabriele Genovese
// @contact.email	gabriele.genovese2@studio.unibo.it
// @license.name	AGPL-3.0
// @license.url	https://www.gnu.org/licenses/agpl-3.0.en.html
// @BasePath		/
func main() {
	err := loadConfig()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	err = util.ConnectDb(config.DbURI)
	if err != nil {
		slog.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}
	db := util.GetDb()
	err = db.AutoMigrate(&proposal.Proposal{}, &api.Question{}, &api.Answer{}, &api.Vote{})
	if err != nil {
		slog.Error("AutoMigrate failed", "err", err)
		os.Exit(1)
	}

	mux := muxie.NewMux()
	mux.Use(util.NewCorsMiddleware(config.ClientURLs, true, mux))

	// authentication-less read-only queries
	mux.HandleFunc("/documents/:id", api.GetDocumentHandler)
	mux.HandleFunc("/questions/:id", api.GetQuestionHandler)

	// authenticated queries
	authMiddleware, err := middleware.NewAuthMiddleware(config.AuthURI)
	if err != nil {
		slog.Error("failed to create authentication middleware", "err", err)
		os.Exit(1)
	}
	mux.Use(authMiddleware.Handler)
	// insert new answer
	mux.HandleFunc("/answers", api.PutAnswerHandler)
	// put up/down votes to an answer
	mux.HandleFunc("/answers/:id/vote", api.PostVote)
	// insert new doc and quesions
	mux.HandleFunc("/documents", api.PutDocumentHandler)

	mux.HandleFunc("/answers/:id", api.DelAnswerHandler)
	// proposal managers
	mux.HandleFunc("/proposals", proposal.ProposalHandler)
	mux.HandleFunc("/proposals/:id", proposal.ProposalByIdHandler)
	mux.HandleFunc("/proposals/document/:id", proposal.ProposalByDocumentHandler)

	slog.Info("listening at", "address", config.Listen)
	err = http.ListenAndServe(config.Listen, mux)
	if err != nil {
		slog.Error("failed to serve", "err", err)
	}
}

func loadConfig() (err error) {
	file, err := os.Open("config.toml")
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	err = toml.NewDecoder(file).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close config file: %w", err)
	}

	return nil
}
