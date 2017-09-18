package server

import (
	"counter"
	"encoding/json"
	"errors"
	"github.com/fatih/structs"
	"github.com/go-pg/pg"
	"io/ioutil"
	"malpar"
	"mylog"
	"net/http"
	"strconv"
	"updater"
)

func getResultId(url string) (int, error) {
	return strconv.Atoi(url[len("/api/result/"):])
}

func (c *Server) FillUserData(results []Result) ([]interface{}, error) {
	ids := make([]interface{}, 0)

	for i := range results {
		for j := range results[i].Pearson {
			current := results[i].Pearson[j]
			ids = append(ids, current.Id)
		}
	}

	var userNames []updater.UserData
	if len(ids) > 0 {
		fields := []string{"id", "name", "last_login", "gender", "birthday", "joined", "location"}
		err := c.db.Model(&userNames).Column(fields...).WhereIn("id IN (?)", ids...).Select()
		if err != nil {
			mylog.Logger.Printf("Get user names: %v", err)
		}
	}

	userNamesMap := make(map[uint32]updater.UserData, 0)
	for i := range userNames {
		userNamesMap[uint32(userNames[i].Id)] = userNames[i]
	}

	filledResults := make([]interface{}, 0)
	for i := range results {
		resultsWithData := make([]interface{}, 0)
		for j := range results[i].Pearson {
			if user, ok := userNamesMap[results[i].Pearson[j].Id]; ok {
				pearsonMap := structs.Map(results[i].Pearson[j])
				pearsonMap["UserName"] = user.Name
				pearsonMap["LastOnline"] = user.LastLogin
				pearsonMap["Gender"] = user.Gender
				pearsonMap["Birthday"] = user.Birthday
				pearsonMap["Joined"] = user.Joined
				pearsonMap["Location"] = user.Location
				resultsWithData = append(resultsWithData, pearsonMap)
			}
		}
		resultMap := structs.Map(results[i])
		resultMap["Pearson"] = resultsWithData
		filledResults = append(filledResults, resultMap)
	}
	return filledResults, nil
}

func (c *Server) listResult(r *http.Request) (interface{}, error) {
	get := r.URL.Query()

	var results []Result
	name, ok := get["username"]
	if ok {
		err := c.db.Model(&results).Where("user_name = ?", name[0]).Order("created desc").Select()
		if err != nil {
			mylog.Logger.Printf("Get result list error: %v", err)
		}
	}

	if results == nil {
		results = make([]Result, 0)
	}

	if len(results) > 0 {
		return c.FillUserData(results)
	}
	return results, nil
}

func (c *Server) createResult(r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return "", errors.New("wrong data")
	}
	receivedResult := Result{}
	err = json.Unmarshal(body, &receivedResult)
	if err != nil {
		return nil, errors.New("wrong data")
	}
	if receivedResult.UserName == "" {
		return nil, errors.New("user name required")
	}
	count, err := c.db.Model(&[]Result{}).Where("user_name = ?", receivedResult.UserName).SelectAndCount()
	if err != pg.ErrNoRows && err != nil {
		mylog.Logger.Printf("Take user results for %v: %v", receivedResult.UserName, err)
	}

	toSaveResult := NewResult()
	toSaveResult.UserName = receivedResult.UserName
	toSaveResult.Anime = receivedResult.Anime
	toSaveResult.Manga = receivedResult.Manga
	toSaveResult.Share = receivedResult.Share
	if count > 20 {
		toSaveResult.Status = COMPLITED
		toSaveResult.Progress = 100
		toSaveResult.Error = "Maximum 20 stored results. Delete some of them."
	}
	err = c.db.Insert(toSaveResult)
	if err != nil {
		mylog.Logger.Printf("Save results for %v: %v", receivedResult.UserName, err)
	}
	if err == nil {
		go func() {
			userList, err := malpar.GetUserScoresByName(toSaveResult.UserName, 3, 0)
			if err != nil {
				toSaveResult.Status = ERROR
				toSaveResult.Error = err.Error()
				err := c.db.Update(toSaveResult)
				if err != nil {
					mylog.Logger.Printf("Update result: %v", err)
				}
			} else {
				user := updater.UserData{Id: uint(userList.UserId), Name: userList.UserName}
				user.UpdateScoresFromList(userList)
				res, err := c.db.Model(&user).Column("anime_scores", "manga_scores").Update()
				if res.RowsAffected() == 0 {
					c.db.Insert(&user)
				}
				if err != nil {
					mylog.Logger.Printf("Update scores %v: %v", user.Name, err)
				}
				c.counter.RequestChan <- counter.ResultRequest{func(pearson counter.PearsonSlice, compared int) {
					toSaveResult.Compare = compared
					toSaveResult.Pearson = pearson
					toSaveResult.Status = COMPLITED
					toSaveResult.Progress = 100
					err := c.db.Update(toSaveResult)
					if err != nil {
						mylog.Logger.Printf("Update result: %v", err)
					}
				}, toSaveResult.Share, toSaveResult.Anime, toSaveResult.Manga, user}
			}
		}()

	}
	return toSaveResult, err
}

func (c *Server) detailResult(r *http.Request) (interface{}, error) {
	resultId, err := getResultId(r.URL.Path)
	result := Result{Id: resultId}
	if err != nil {
		return result, err
	}
	err = c.db.Select(&result)
	if err != nil {
		return result, err
	}
	filledResults, err := c.FillUserData([]Result{result})
	return filledResults[0], err
}

func (c *Server) deleteResult(r *http.Request) (interface{}, error) {
	resultId, err := getResultId(r.URL.Path)
	if err != nil {
		return nil, err
	}
	_, err = c.db.Model(&Result{}).Where("id = ?", resultId).Delete()
	return nil, err
}

func (c *Server) serveResult(w http.ResponseWriter, r *http.Request) {
	var res interface{}
	var err error

	switch r.Method {
	case "GET":
		strId := r.URL.Path[len("/api/result/"):]
		if len(strId) == 0 {
			res, err = c.listResult(r)
		} else {
			res, err = c.detailResult(r)
		}
	case "POST":
		res, err = c.createResult(r)
	case "DELETE":
		res, err = c.deleteResult(r)

	}

	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(FormatResponse(res, err))
}
