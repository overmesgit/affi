package counter

import (
	"fmt"
	"github.com/go-pg/pg"
	"mylog"
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
	lastLoadedId := 0
	scores := 0
	for i := 0; i < 5; i++ {
		var users []updater.UserData
		err := p.db.Model(&users).Where("id > ?", lastLoadedId).Order("id ASC").Limit(100000).Select()
		if len(users) == 0 {
			break
		}
		if err != nil {
			mylog.Logger.Panicf("Load user scores: %v", err)
		} else {
			for i := range users {
				p.Pearson.UpdateUserSlices(users[i])
				lastLoadedId = users[i].Id
				scores += len(users[i].AnimeScores) + len(users[i].MangaScores)
			}
		}
		fmt.Printf("Loaded user before id %v\n", lastLoadedId)
		fmt.Printf("Loaded scores %v\n", scores)
	}
	mylog.Logger.Printf("Loaded users %v", len(p.Pearson.AnimeIndexes))
	mylog.Logger.Printf("Scores: %v", scores)
	mylog.Logger.Printf("Anime Items: %v", len(p.Pearson.AnimeIdReplace))
	mylog.Logger.Printf("Manga Items: %v", len(p.Pearson.MangaIdReplace))
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
