package api

import (
	"encoding/json"
	"net/http"

	"github.com/csunibo/polleg/auth"
	"github.com/csunibo/polleg/util"
	"github.com/kataras/muxie"
)

type AnswerObj struct {
	Question uint   `json:"question"`
	Parent   *uint  `json:"parent"`
	Content  string `json:"content"`
}

// Insert a new answer under a question
func PutAnswerHandler(res http.ResponseWriter, req *http.Request) {
	// Check method PUT is used
	if req.Method != http.MethodPut {
		_ = util.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	user := auth.GetUser(req)

	// Declare a new Person struct.
	var ans AnswerObj
	err := json.NewDecoder(req.Body).Decode(&ans)
	if err != nil {
		util.WriteError(res, http.StatusBadRequest, "decode error")
		return
	}

	var quest Question
	if err := db.First(&quest, ans.Question).Error; err != nil {
		util.WriteError(res, http.StatusBadRequest, "no Question associated with request (or other Error)")
		return
	}

	if ans.Parent != nil {
		var Parent Answer
		if err = db.First(&Parent, ans.Parent).Error; err != nil {
			util.WriteError(res, http.StatusBadRequest, "parent is given but none found")
			return
		}
		if Parent.Question != quest.ID {
			util.WriteError(res, http.StatusBadRequest, "mismatch between parent question and this question")
			return
		}
	}

	err = db.Create(&Answer{
		Question:  ans.Question,
		Parent:    ans.Parent,
		User:      user.Username,
		Content:   ans.Content,
		Upvotes:   0,
		Downvotes: 0,
	}).Error

	if err != nil {
		util.WriteError(res, http.StatusBadRequest, "create error")
		return
	}

	if err = util.WriteJson(res, util.Res{Res: "OK"}); err != nil {
		util.WriteError(res, http.StatusInternalServerError, "couldn't write response")
	}
}

// Get an answer by an ID
func GetAnswerById(res http.ResponseWriter, req *http.Request) {
	db := util.GetDb()
	id := muxie.GetParam(res, "id")

	var ans Answer
	if err := db.First(&ans, id).Error; err != nil {
		util.WriteError(res, http.StatusBadRequest, "Answer not found")
		return
	}

	if err := util.WriteJson(res, ans); err != nil {
		util.WriteError(res, http.StatusInternalServerError, "couldn't write response")
	}
}

// Given a question ID, find all the answers
func GetAnswersByQuestion(res http.ResponseWriter, req *http.Request) {
	db := util.GetDb()
	qid := muxie.GetParam(res, "id")

	var ans []Answer
	if err := db.Where("question = ?", qid).Find(&ans).Error; err != nil {
		util.WriteError(res, http.StatusInternalServerError, "Answer not found")
		return
	}

	if err := util.WriteJson(res, ans); err != nil {
		util.WriteError(res, http.StatusInternalServerError, "couldn't write response")
	}
}