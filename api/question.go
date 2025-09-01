package api

import (
	"net/http"
	"strconv"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
)

// @Summary		Delete a question
// @Description	Given an andwer ID, delete the question
// @Tags			question
// @Param			id	path	string	true	"Question id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/questions/{id} [delete]
func DelQuestionHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	user := middleware.GetUser(req)
	if !user.Admin {
		httputil.WriteError(res, http.StatusForbidden, "only admins can delete questions")
		return
	}

	db := util.GetDb()
	rawAnsID := muxie.GetParam(res, "id")

	qID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	if err := db.Delete(&models.Question{}, uint(qID)).Error; err != nil {
		slog.Error("something went wrong", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "something went wrong")
		return
	}

	httputil.WriteData(res, http.StatusNoContent, nil)
}
