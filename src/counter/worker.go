package counter

import (
	"fmt"
	"github.com/go-pg/pg"
	"mylog"
	"runtime"
	"time"
	"updater"
)

type ResultRequest struct {
	OnSuccess func(pearson PearsonSlice, compared int)
	Share     int
	Anime     bool
	Manga     bool
	User      updater.UserData
}

type PearsonCounter struct {
	db          *pg.DB
	Pearson     *Pearson
	RequestChan chan ResultRequest
	UpdateChan  chan updater.UserData
}

func NewPearsonCounter(db *pg.DB) *PearsonCounter {
	return &PearsonCounter{db: db, Pearson: NewPearson(),
		RequestChan: make(chan ResultRequest, 100), UpdateChan: make(chan updater.UserData, 100)}
}

func (p *PearsonCounter) Prepare() {
	limit := 10000
	scores := 0
	lastLoadedId := uint(0)
	for {
		startTime := time.Now().Unix()
		var users []updater.UserData
		err := p.db.Model(&users).Where("id > ?", lastLoadedId).Order("id ASC").Limit(limit).Select()
		if len(users) == 0 {
			break
		}
		if err != nil {
			mylog.Logger.Panicf("Load user scores: %v", err)
		} else {
			for i := range users {
				if len(users[i].AnimeScores)+len(users[i].MangaScores) > 10 {
					p.Pearson.UpdateUserSlices(users[i])
					scores += len(users[i].AnimeScores) + len(users[i].MangaScores)
				}
				lastLoadedId = users[i].Id
			}
		}
		fmt.Printf("Loaded user before id:	%v, scores:	%v, time:	%v\n", lastLoadedId, scores, time.Now().Unix()-startTime)
		runtime.GC()
	}
	mylog.Logger.Printf("Loaded users %v", len(p.Pearson.AnimeIndexes))
	mylog.Logger.Printf("Scores: %v", scores)
	mylog.Logger.Printf("Anime Items: %v", len(p.Pearson.AnimeIdToIndex))
	mylog.Logger.Printf("Manga Items: %v", len(p.Pearson.MangaIdToIndex))
}

func (p *PearsonCounter) Start() {
	go func() {
		for {
			select {
			case list := <-p.UpdateChan:
				p.Pearson.UpdateUserSlices(list)
			case req := <-p.RequestChan:
				pearson, compare := p.Pearson.Count(req.User, req.Share, req.Anime, req.Manga)
				go func() {
					req.OnSuccess(pearson, compare)
				}()
			}
		}
	}()
}
