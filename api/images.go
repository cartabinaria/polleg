package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/google/uuid"
	"github.com/kataras/muxie"
)

type ImageType string

var (
	// File signatures (magic numbers)
	pngSignature  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpegSignature = []byte{0xFF, 0xD8, 0xFF}

	ImageTypePNG  ImageType = "image/png"
	ImageTypeJPEG ImageType = "image/jpeg"
)

const (
	MAX_IMAGE_SIZE = 5 * 1024 * 1024 // 5 MB
)

// checkFileType reads the first few bytes of a file and compares them with known signatures
func checkFileType(reader io.Reader) (ImageType, error) {
	// Read first 8 bytes for signature checking
	buff := make([]byte, 8)
	n, err := reader.Read(buff)
	if err != nil || n < 8 {
		return "", fmt.Errorf("error reading file header: %v", err)
	}

	// Check signatures
	if bytes.HasPrefix(buff, pngSignature) {
		return ImageTypePNG, nil
	}
	if bytes.HasPrefix(buff, jpegSignature) {
		return ImageTypeJPEG, nil
	}

	return "", fmt.Errorf("unsupported file type")
}

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

		if fpCheck, err := checkFileType(file); err != nil {
			slog.With("err", err).Error("couldn't check file type")
			http.Error(w, "couldn't check file type", http.StatusBadRequest)
			return
		} else if string(fpCheck) != fType {
			slog.With("expected", fType, "got", fpCheck).Error("file type mismatch")
			http.Error(w, "file type mismatch", http.StatusBadRequest)
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
		switch {
		case err == io.EOF:
			// File is within size limits - this is good!
			slog.With("path", fullPath, "size", written).Info("file successfully saved")
		case err != nil:
			// Unexpected error occurred
			slog.With("err", err).Error("couldn't save file")
			if cleanupErr := os.Remove(fullPath); cleanupErr != nil {
				slog.With("err", cleanupErr, "path", fullPath).Error("couldn't remove file after failed save")
			}
			http.Error(w, "couldn't save file", http.StatusInternalServerError)
			return
		case written > MAX_IMAGE_SIZE:
			// File exceeded size limit
			slog.With("size", written, "max", MAX_IMAGE_SIZE).Error("file too large")
			if cleanupErr := os.Remove(fullPath); cleanupErr != nil {
				slog.With("err", cleanupErr, "path", fullPath).Error("couldn't remove file after failed save")
			}
			http.Error(w, "file too large", http.StatusBadRequest)
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

		userID := middleware.GetUser(r).ID

		db := util.GetDb()
		_, err = util.CreateImage(db, uuid.String(), userID, uint(written))
		if err != nil {
			slog.With("err", err).Error("couldn't create image record")
			if cleanupErr := os.Remove(fullPath); cleanupErr != nil {
				slog.With("err", cleanupErr, "path", fullPath).Error("couldn't remove file after failed db record creation")
			}
			http.Error(w, "couldn't create image record", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ImageResponse{
			ID:  uuid.String(),
			URL: fullPath,
		})
	}
}
