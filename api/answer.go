package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
    "regexp"
    "math/rand"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

type PutAnswerRequest struct {
	Question  uint   `json:"question"`
	Parent    *uint  `json:"parent"`
	Content   string `json:"content"`
	Anonymous bool   `json:"anonymous"`
}

type AnswerResponse struct {
	// taken from from gorm.Model, so we can json strigify properly
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Question uint  `json:"question"`
	Parent   *uint `json:"parent"`

	User      string   `json:"user"`
	Content   string   `json:"content"`
	Upvotes   uint32   `json:"upvotes"`
	Downvotes uint32   `json:"downvotes"`
	Replies   []Answer `json:"replies"`

	Anonymous bool     `json:"anonymous"`
    AnonymousAvatarURL string `json:"anonymous_avatar_url"`
    Alias string `json:"alias"`
}

var (
	VOTES_QUERY = fmt.Sprintf(`
  select   votes.answer,
           count(case votes.vote when %d then 1 else null end) as upvotes,
           count(case votes.vote when %d then 1 else null end) as downvotes
  from     votes
  where deleted_at is NULL
  group by answer
`, VoteUp, VoteDown)
	ANSWERS_QUERY = fmt.Sprintf(`
  select   *
  from     answers
  full join     (%s) on answer = answers.id
  where    deleted_at is NULL and answers.parent is NULL and answers.question = ?
`, VOTES_QUERY)
	REPLIES_QUERY = `
  select   *
  from     answers
  where    deleted_at is NULL and answers.parent IN ?
`
	names = []string{"Wyatt", "Vivian", "Maria", "Alexander", "Luis", "Aidan", "Mason", "Aiden", "Mackenzie", "Adrian", "Oliver", "Andrea", "Amaya", "Nolan", "Riley", "Robert", "Ryker", "Sara", "Ryan", "Sawyer"}
)

func GetOrCreateUserByID(db *gorm.DB, id uint, username string) (*User, error) {

    // Try to find user by ID
    var user User
    userResult := db.First(&user, id)
    if userResult.Error != nil {
        if userResult.Error == gorm.ErrRecordNotFound {
            name := names[rand.Intn(len(names))]

            // Find last alias for this name
            var lastUser User
            pattern := fmt.Sprintf("%s%%", name)
            result := db.Where("alias LIKE ?", pattern).Order("created_at DESC").First(&lastUser)

            // Extract last number
            nextNum := 1
            if result.Error == nil {
                re := regexp.MustCompile(fmt.Sprintf(`^%s(\d+)$`, name))
                matches := re.FindStringSubmatch(lastUser.Alias)
                if len(matches) == 2 {
                    if n, err := strconv.Atoi(matches[1]); err == nil {
                        nextNum = n + 1
                    }
                }
            }

            alias := fmt.Sprintf("%s%d", name, nextNum)

            // Not found, create new record
            user = User{
                ID:       id,
                Username: username,
                Alias:    alias,
            }
            if err := db.Create(&user).Error; err != nil {
                return nil, err
            }
            return &user, nil
        }
        // Some other error
        return nil, userResult.Error
    }
    // Found
    return &user, nil
}

func generateAnonymousAvatar(alias string) string {
    re := regexp.MustCompile(`^[A-Za-z]+`)
    match := re.FindString(alias)
    if match == "" {
        return ""
    }
    return fmt.Sprintf("https://api.dicebear.com/9.x/thumbs/svg?seed=%s", match)
}

// @Summary		Insert a new answer
// @Description	Insert a new answer under a question
// @Tags			answer
// @Param			answerReq	body	PutAnswerRequest	true	"Answer data to insert"
// @Produce		json
// @Success		200	{object}	Answer
// @Failure		400	{object}	util.ApiError
// @Router			/answers [put]
func PutAnswerHandler(res http.ResponseWriter, req *http.Request) {
	// Check method PUT is used
	if req.Method != http.MethodPut {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	user := middleware.GetUser(req)

	var ans PutAnswerRequest
	err := json.NewDecoder(req.Body).Decode(&ans)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, fmt.Sprintf("decode error: %v", err))
		return
	}

	var quest Question
	if err := db.First(&quest, ans.Question).Error; err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "the referenced question does not exist")
		return
	}

	if ans.Parent != nil {
		var Parent Answer
		if err = db.First(&Parent, ans.Parent).Error; err != nil {
			httputil.WriteError(res, http.StatusBadRequest, "the referenced parent does not exist")
			return
		}
		if Parent.Question != quest.ID {
			httputil.WriteError(res, http.StatusBadRequest, "mismatch between parent question and this question")
			return
		}
	}

	// TODO: upvotes and downvotes should really be just the result of a
	// COUNT() aggregator on the votes table
	answer := Answer{
		Question:  ans.Question,
		Parent:    ans.Parent,
		UserId:    user.ID,
		Content:   ans.Content,
		Upvotes:   0,
		Downvotes: 0,
		Anonymous: ans.Anonymous,
	}

	err = db.Create(&answer).Error
	if err != nil {
		slog.Error("error while creating the answer", "answer", answer, "err", err)
		httputil.WriteError(res, http.StatusBadRequest, "could not insert the answer")
		return
	}

    usr, err := GetOrCreateUserByID(db, user.ID, user.Username)
    if err != nil {
        slog.Error("error while getting or creating the user-alias association", "user", user, "err", err)
        httputil.WriteError(res, http.StatusBadRequest, "could not insert the answer")
        return
    }
    
	httputil.WriteData(res, http.StatusOK, AnswerResponse{
		ID:        answer.ID,
		CreatedAt: answer.CreatedAt,
		UpdatedAt: answer.UpdatedAt,
		Question:  answer.Question,
		Parent:    answer.Parent,
		User:      user.Username,
		Content:   answer.Content,
		Upvotes:   answer.Upvotes,
		Downvotes: answer.Downvotes,
		Replies:   answer.Replies,
		Anonymous: answer.Anonymous,
        AnonymousAvatarURL:    generateAnonymousAvatar(usr.Alias),
        Alias: usr.Alias,
	})
}

// @Summary		Get all answers given a question
// @Description	Given a question ID, return the question and all its answers
// @Tags			question
// @Param			id	path	string	true	"Answer id"
// @Produce		json
// @Success		200	{array}		Answer
// @Failure		400	{object}	util.ApiError
// @Router			/questions/{id} [get]
func GetQuestionHandler(res http.ResponseWriter, req *http.Request) {
	// Check method GET is used
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	rawQID := muxie.GetParam(res, "id")
	qID, err := strconv.ParseUint(rawQID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	var question Question
	if err := db.First(&question, uint(qID)).Error; err != nil {
		slog.Error("question not found", "err", err)
		httputil.WriteError(res, http.StatusNotFound, "question not found")
		return
	}
	var answers []Answer
	if err := db.Raw(ANSWERS_QUERY, question.ID).Scan(&answers).Error; err != nil {
		slog.Error("could not fetch answers", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "could not fetch answers")
		return
	}
	answersIDs := []uint{}
	answersIndex := map[uint]int{}
	for i, answer := range answers {
		answersIDs = append(answersIDs, answer.ID)
		answersIndex[answer.ID] = i
	}
	var replies []Answer
	if err := db.Raw(REPLIES_QUERY, answersIDs).Scan(&replies).Error; err != nil {
		slog.Error("could not fetch replies", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "could not fetch replies")
		return
	}
	for _, reply := range replies {
		index := answersIndex[*reply.Parent]
		answers[index].Replies = append(answers[index].Replies, reply)
	}

	question.Answers = answers
	httputil.WriteData(res, http.StatusOK, question)
}

// @Summary		Delete an answer
// @Description	Given an andwer ID, delete the answer
// @Tags			answer
// @Param			id	path	string	true	"Answer id"
// @Produce		json
// @Success		200	{object}	Answer
// @Failure		400	{object}	util.ApiError
// @Router			/answers/{id} [delete]
func DelAnswerHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	user := middleware.GetUser(req)
	db := util.GetDb()
	rawAnsID := muxie.GetParam(res, "id")

	aID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	var ans Answer
	if err := db.First(&ans, uint(aID)).Error; err != nil {
		slog.Error("answer not found", "err", err)
		httputil.WriteError(res, http.StatusNotFound, "answer not found")
		return
	}

	if !user.Admin && ans.UserId != user.ID {
		slog.Error("you are not an admin or the owner of the answer", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "you are not an admin or the owner of the answer")
		return
	}

	if err := db.Delete(&ans).Error; err != nil {
		slog.Error("something went wrong", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "something went wrong")
		return
	}

	httputil.WriteData(res, http.StatusOK, AnswerResponse{
		ID:        ans.ID,
		CreatedAt: ans.CreatedAt,
		UpdatedAt: ans.UpdatedAt,
		Question:  ans.Question,
		Parent:    ans.Parent,
		User:      user.Username,
		Content:   ans.Content,
		Upvotes:   ans.Upvotes,
		Downvotes: ans.Downvotes,
		Replies:   ans.Replies,
		Anonymous: ans.Anonymous,
	})
}
