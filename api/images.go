package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cartabinaria/polleg/models"
	"github.com/google/uuid"
	"github.com/kataras/muxie"
)

const (
	MAX_IMAGE_SIZE = 5 * 1024 * 1024 // 5 MB
)

// @Summary		Get an image
// @Description	Given an image ID, return the image
// @Tags			image
// @Param			id	path	string	true	"Image id"
// @Produce		json
// @Success		200	{file}		binary
// @Failure		400	{object}	httputil.ApiError
// @Router			/questions/{id} [get]
func GetImageHandler(imagesPath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		imgID := muxie.GetParam(res, "id")

		_, err := uuid.Parse(imgID)
		if err != nil {
			http.Error(res, "invalid image id", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(imagesPath, imgID)

		http.ServeFile(res, req, fullPath)
	}
}

// @Summary		Insert a new image
// @Description	Insert a new image
// @Tags			image
// @Accept			multipart/form-data
// @Param			image	formData	file	true	"Image to upload"
// @Produce		json
// @Success		200	{object}	models.ImageResponse
// @Failure		400	{object}	httputil.ApiError
// @Router			/images [post]
func PostImageHandler(imagesPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			slog.With("err", err).Error("couldn't get file from form")
			http.Error(w, "couldn't get file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		slog.With("filename", fileHeader.Filename, "size", fileHeader.Size, "Type: ", fileHeader.Header.Get("Content-Type")).Info("received file")

		if fileHeader.Size > MAX_IMAGE_SIZE {
			http.Error(w, "file too large", http.StatusBadRequest)
			return
		}

		fType := fileHeader.Header.Get("Content-Type")
		if fType != "image/png" && fType != "image/jpeg" {
			http.Error(w, "invalid file type", http.StatusBadRequest)
			return
		}

		uuid, err := uuid.NewV7()
		if err != nil {
			slog.With("err", err).Error("couldn't generate uuid")
			http.Error(w, "couldn't generate uuid", http.StatusInternalServerError)
			return
		}
		fullPath := filepath.Join(imagesPath, uuid.String())

		destFile, err := os.Create(fullPath)
		if err != nil {
			slog.With("err", err).Error("couldn't create file")
			http.Error(w, "couldn't create file", http.StatusInternalServerError)
			return
		}
		defer destFile.Close()

		written, err := io.CopyN(destFile, file, MAX_IMAGE_SIZE+1)
		if err != nil && err != io.EOF {
			slog.With("err", err).Error("couldn't save file")
			err = os.Remove(fullPath)
			if err != nil {
				slog.With("err", err, "path", fullPath).Error("couldn't remove file after failed save")
			}
			http.Error(w, "couldn't save file", http.StatusInternalServerError)
			return
		}

		if written > MAX_IMAGE_SIZE {
			err = os.Remove(fullPath)
			if err != nil {
				slog.With("err", err, "path", fullPath).Error("couldn't remove file after failed save")
			}
			http.Error(w, "file too large", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ImageResponse{
			ID:  uuid.String(),
			URL: fullPath,
		})
	}
}
