package proposal

import (
	"net/http"
	"strconv"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
)

func ProposalByIdHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}

	switch req.Method {
	case http.MethodDelete:
		deleteProposalByIdHandler(res)
	case http.MethodGet:
		getProposalByIdHandler(res)
	default:
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
	}
}

func getProposalByIdHandler(res http.ResponseWriter) {
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

func deleteProposalByIdHandler(res http.ResponseWriter) {
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
