package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kataras/muxie"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/slog"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/api"
	"github.com/cartabinaria/polleg/api/proposal"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
)

type Config struct {
	Listen     string   `toml:"listen"`
	ClientURLs []string `toml:"client_urls"`

	DbURI   string `toml:"db_uri" required:"true"`
	AuthURI string `toml:"auth_uri" required:"true"`

	ImagesPath string `toml:"images_path"`
}

var (
	// Default config values
	config = Config{
		Listen:     "0.0.0.0:3001",
		AuthURI:    "http://localhost:3000",
		ImagesPath: "./images",
	}
)

// @title			Polleg API
// @version		1.0
// @description	This is the backend API for Polleg that allows unibo students to answer exam exercises directly on the cartabinaria website
// @contact.name	Gabriele Genovese
// @contact.email	gabriele.genovese2@studio.unibo.it
// @license.name	AGPL-3.0
// @license.url	https://www.gnu.org/licenses/agpl-3.0.en.html
// @BasePath		/
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: polleg <config-file>")
		os.Exit(1)
	}
	err := loadConfig(os.Args[1])
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
	err = db.AutoMigrate(&models.User{}, &models.Proposal{}, &models.Question{}, &models.Answer{}, &models.Vote{}, &models.Image{}, &models.AnswerVersion{}, &models.Report{})
	if err != nil {
		slog.Error("AutoMigrate failed", "err", err)
		os.Exit(1)
	}

	err = os.Mkdir(config.ImagesPath, 0755)
	if err != nil && !os.IsExist(err) {
		slog.Error("failed to create images directory", "err", err)
		os.Exit(1)
	}

	mux := muxie.NewMux()
	authMiddleware, err := middleware.NewAuthMiddleware(config.AuthURI)
	if err != nil {
		slog.Error("failed to create authentication middleware", "err", err)
		os.Exit(1)
	}

	mux.Use(util.NewLoggerMiddleware, httputil.NewCorsMiddleware(config.ClientURLs, true, mux))

	authChain := muxie.Pre(authMiddleware.Handler, api.BanMiddleware)
	authOptionalChain := muxie.Pre(authMiddleware.NonBlockingHandler)

	// authentication-less read-only queries
	mux.Handle("/documents", authOptionalChain.ForFunc(api.GetDocumentsWithQuestionsHandler))
	mux.Handle("/documents/:id", authOptionalChain.ForFunc(api.GetDocumentHandler))
	mux.Handle("/questions/:id", muxie.Methods().
		Handle("GET", authOptionalChain.ForFunc(api.GetQuestionHandler)).
		Handle("DELETE", authChain.ForFunc(api.DelQuestionHandler)))

	mux.Handle("/images/:id", authOptionalChain.ForFunc(api.GetImageHandler(config.ImagesPath)))

	// authenticated queries
	// insert new answer
	mux.Handle("/answers", authChain.ForFunc(api.PostAnswerHandler))
	// put up/down votes to an answer
	mux.Handle("/answers/:id/vote", authChain.ForFunc(api.PostVote))
	mux.Handle("/answers/:id/replies", authOptionalChain.ForFunc(api.GetRepliesHandler))
	// insert new doc and quesions
	mux.Handle("/documents", authChain.ForFunc(api.PostDocumentHandler))
	mux.Handle("/answers/:id", authChain.ForFunc(api.DelAnswerHandler))
	mux.Handle("/answers/:id", muxie.Methods().
		Handle("DELETE", authChain.ForFunc(api.DelAnswerHandler)).
		Handle("PATCH", authChain.ForFunc(api.UpdateAnswerHandler)))

	// Images
	mux.Handle("/images", authChain.ForFunc(api.PostImageHandler(config.ImagesPath)))

	// proposal managers
	mux.Handle("/proposals", muxie.Methods().
		Handle("POST", authChain.ForFunc(proposal.PostProposalHandler)).
		Handle("GET", authChain.ForFunc(proposal.GetAllProposalsHandler)))
	mux.Handle("/proposals/:id/approve", authChain.ForFunc(proposal.ApproveProposalHandler))
	mux.Handle("/proposals/:id", muxie.Methods().
		Handle("DELETE", authChain.ForFunc(proposal.DeleteProposalByIdHandler)).
		Handle("GET", authChain.ForFunc(proposal.GetProposalByIdHandler)))
	mux.Handle("/proposals/document/:id", muxie.Methods().
		Handle("GET", authChain.ForFunc(proposal.GetProposalByDocumentHandler)).
		Handle("DELETE", authChain.ForFunc(proposal.DeleteProposalByDocumentHandler)))
	mux.Handle("/proposals/document/:id/approve", authChain.ForFunc(proposal.ApproveProposalByDocumentHandler))

	// Logs
	mux.Handle("/logs", authChain.ForFunc(api.LogsHandler))

	// Moderation
	mux.Handle("/moderation/report/:id", authChain.ForFunc(api.ReportByIdHandler))
	mux.Handle("/moderation/reports", authChain.ForFunc(api.GetReportsHandler))
	mux.Handle("/moderation/ban", muxie.Methods().
		Handle("GET", authChain.ForFunc(api.GetBannedHandler)).
		Handle("POST", authChain.ForFunc(api.BanUserHandler)))

	// start garbage collector
	go util.GarbageCollector(config.ImagesPath)

	slog.Info("listening at", "address", config.Listen)
	err = http.ListenAndServe(config.Listen, mux)
	if err != nil {
		slog.Error("failed to serve", "err", err)
	}
}

func loadConfig(path string) (err error) {
	file, err := os.Open(path)
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
