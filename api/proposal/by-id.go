package proposal

import (
	"net/http"
	"strconv"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
)

// @Summary		Get proposal by id
// @Description	Get a proposal given its ID
// @Tags			proposal
// @Param			id	path	string	true	"Proposal id"
// @Produce		json
// @Success		200	{object}	Proposal
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals/{id} [get]
func GetProposalByIdHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}

	db := util.GetDb()
	proposalID := muxie.GetParam(res, "id")
	propID, err := strconv.ParseUint(proposalID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid answer id")
		return
	}

	var props Proposal
	if err := db.Where(Proposal{ID: propID}).Take(&props).Error; err != nil {
		httputil.WriteError(res, http.StatusNotFound, "Not found")
		return
	}

	httputil.WriteData(res, http.StatusOK, props)
}

// @Summary		Delete a proposal
// @Description	Given a proposal ID, delete the proposal
// @Tags			proposal
// @Param			id	path	string	true	"Proposal id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals/{id} [delete]
func DeleteProposalByIdHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}

	db := util.GetDb()
	proposalID := muxie.GetParam(res, "id")
	propID, err := strconv.ParseUint(proposalID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid answer id")
		return
	}

	if err := db.Delete(&Proposal{}, propID).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}

	res.WriteHeader(http.StatusNoContent)
}
