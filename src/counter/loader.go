package counter

import (
	"github.com/go-pg/pg"
	"mylog"
	"updater"
)

//type RequestChannel struct {
//	control.PearsonRequest
//	Response chan malpar.PearsonSlice
//}

type PearsonCounter struct {
	db      *pg.DB
	Pearson *Pearson
	//RequestChan chan RequestChannel
	//UpdateChan chan malpar.UserList
}

func NewPearsonCounter(db *pg.DB) *PearsonCounter {
	return &PearsonCounter{db, NewPearson()}
}

//func NewPearsonCounter() *PearsonCounter {
//	return &PearsonCounter{Pearson: malpar.NewPearson(), RequestChan: make(chan RequestChannel, 10), UpdateChan: make(chan malpar.UserList, 1000)}
//}
//
//func (p *PearsonCounter) UpdateScores(list *malpar.UserList) {
//	if countServer.IsMyUser(list.UserId) {
//		p.UpdateChan <- *list
//	}
//}

func (p *PearsonCounter) Prepare() {
	lastLoadedId := 0
	for {
		var users []updater.UserData
		err := p.db.Model(&users).Where("id > ?", lastLoadedId).Order("id ASC").Limit(1000).Select()
		if len(users) == 0 {
			break
		}
		if err != nil {
			mylog.Logger.Panicf("Load user scores: %v", err)
		} else {
			for i := range users {
				p.Pearson.UpdateUserSlices(users[i])
				lastLoadedId = users[i].Id
			}
		}
	}
	mylog.Logger.Printf("Loaded users %v", len(p.Pearson.AnimeIndexes))
	mylog.Logger.Printf("Anime Items: %v", len(p.Pearson.AnimeIdReplace))
	mylog.Logger.Printf("Manga Items: %v", len(p.Pearson.MangaIdReplace))
}

//func (p *PearsonCounter) Start() {
//	//var list malpar.UserList
//	for {
//		select {
//		case list := <-p.UpdateChan:
//			if countServer.IsMyUser(list.UserId) {
//				p.Pearson.UpdateUserList(list)
//			}
//		case req := <-p.RequestChan:
//			pearson := p.Pearson.Count(req.UserList, req.Share, req.Anime, req.Manga, countServer.IsMyUser(req.UserList.UserId))
//			req.Response <- pearson
//		}
//	}
//}
