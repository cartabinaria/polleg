package util

import (
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"time"

	"github.com/cartabinaria/polleg/models"
)

// Images are uploaded before being posted in a answer, so we could end
// up with unused images. This garbage collector removes images that are in the
// database for more than 24 hours but not attached to any question. To check
// if an image is attached to a answer, we check if its URL is present in
// the Content field of any answer.

func GarbageCollector(imagesPath string) {
	slog.Info("starting garbage collector")
	ticker := time.NewTicker(24 * time.Hour)

	for range ticker.C {
		slog.Info("running garbage collector")
		if err := cleanUnusedImages(imagesPath); err != nil {
			slog.With("err", err).Error("error while cleaning unused images")
		}
	}
}

func cleanUnusedImages(imagesPath string) error {
	cutoff := time.Now().Add(-24 * time.Hour)
	db := GetDb()

	var oldImages []models.Image
	if err := db.Where("created_at < ?", cutoff).Find(&oldImages).Error; err != nil {
		return err
	}

	var answersContent []string
	err := db.Table("answer_versions av1").
		Select("av1.content").
		Joins("INNER JOIN (SELECT answer, MAX(id) as max_id FROM answer_versions GROUP BY answer) av2 ON av1.answer = av2.answer AND av1.id = av2.max_id").
		Pluck("content", &answersContent).Error

	if err != nil {
		return err
	}

	for _, img := range oldImages {
		// Maybe this regex is too much, we'll see
		containsRegex := regexp.MustCompile(`\!\[.*\]\(\s*https:\/\/[^\s]+\/images\/` + img.ID + `.*\)`)

		if slices.ContainsFunc(answersContent, func(content string) bool {
			return containsRegex.MatchString(content)
		}) {
			continue
		}

		err := os.Remove(filepath.Join(imagesPath, img.ID))
		if err != nil && !os.IsNotExist(err) {
			slog.With("image", img, "err", err).Error("error while deleting unused image file")
		}
		if err := db.Delete(&img).Error; err != nil {
			slog.With("image", img, "err", err).Error("error while deleting unused image")
			continue
		}
		slog.With("image", img).Info("deleted unused image")
	}

	return nil
}
