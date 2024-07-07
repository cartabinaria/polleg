package api

import (
	"encoding/json"
	"net/http"

	"github.com/csunibo/auth/pkg/httputil"
	"github.com/csunibo/auth/pkg/middleware"
	"github.com/csunibo/polleg/util"
	"github.com/kataras/muxie"
)

type Document struct {
	ID        string     `json:"id"`
	Questions []Question `json:"questions"`
}

type Coord struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

type PutDocumentRequest struct {
	ID     string  `json:"id"`
	Coords []Coord `json:"coords"`
}

// @Summary		Insert a new document
// @Description	Insert a new document with all the questions initialised
// @Tags			document
// @Param			docRequest	body	PutDocumentRequest	true	"Doc request body"
// @Produce		json
// @Success		200	{object}	Document
// @Failure		400	{object}	util.ApiError
// @Router			/documents [put]
func PutDocumentHandler(res http.ResponseWriter, req *http.Request) {
	// only members of the staff can add a document
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}
	// Check method PUT is used
	if req.Method != http.MethodPut {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()

	// decode data
	var data PutDocumentRequest
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "couldn't decode body")
		return
	}

	// save questions
	var questions []Question
	for _, coord := range data.Coords {
		q := Question{
			Document: data.ID,
			Start:    coord.Start,
			End:      coord.End,
		}
		questions = append(questions, q)
	}

	if err := db.Save(questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't create questions")
		return
	}

	httputil.WriteData(res, http.StatusOK, Document{
		ID:        data.ID,
		Questions: questions,
	})
}

// @Summary		Get a document's divisions
// @Description	Given a document's ID, return all the questions
// @Tags			document
// @Param			id	path	string	true	"document id"
// @Produce		json
// @Success		200	{object}	Document
// @Failure		400	{object}	util.ApiError
// @Router			/documents/{id} [get]
func GetDocumentHandler(res http.ResponseWriter, req *http.Request) {
	// Check method GET is used
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	docID := muxie.GetParam(res, "id")
	var questions []Question
	if err := db.Where(Question{Document: docID}).Find(&questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}
	if len(questions) == 0 {
		httputil.WriteError(res, http.StatusNotFound, "Document not found")
		return
	}
	httputil.WriteData(res, http.StatusOK, Document{
		ID:        docID,
		Questions: questions,
	})
}
