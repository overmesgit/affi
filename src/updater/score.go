package updater

import (
	"github.com/go-pg/pg"
	"malpar"
	"mylog"
)

type ScoreUpdater struct {
	db       *pg.DB
	toUpdate chan UserData
	onUpdate func(data UserData)
}

func NewScoreUpdater(db *pg.DB, onUpdate func(data UserData)) *ScoreUpdater {
	return &ScoreUpdater{db: db, toUpdate: make(chan UserData, 1), onUpdate: onUpdate}
}

// Concurrently update all user scores in db
func (s *ScoreUpdater) Start(threads int) {
	go s.fillUpdateQueue()
	for i := 0; i < threads; i++ {
		go s.startThread()
	}

}

// Get last updated id and start fill queue with users which id > than taken
// After full update start from 0
func (s *ScoreUpdater) fillUpdateQueue() {
	lastUpdate, err := GetLastUpdate(s.db)
	if err != nil {
		mylog.Logger.Printf("Select last update id: %v", err)
	}

	for {
		var nextUser UserData
		err := s.db.Model(&nextUser).Where("id > ?", lastUpdate.LastScoreId).Order("id asc").First()
		if err == pg.ErrNoRows {
			mylog.Logger.Printf("User score updated, restart, last id: %v", lastUpdate.LastScoreId)
			lastUpdate.LastScoreId = 0
			_, err = s.db.Model(&lastUpdate).Column("last_score_id").Update()
			if err != nil {
				mylog.Logger.Printf("Update LastUpdate: %v", err)
			}
			continue
		}

		if err != nil {
			mylog.Logger.Printf("Get next user: %v", err)
			continue
		}

		s.toUpdate <- nextUser

		lastUpdate.LastScoreId = nextUser.Id
		_, err = s.db.Model(&lastUpdate).Column("last_score_id").Update()
		if err != nil {
			mylog.Logger.Printf("Update Scores LastUpdate: %v", err)
		}
	}
}

func (s *ScoreUpdater) startThread() {
	for {
		user := <-s.toUpdate
		mylog.Logger.Printf("Get score for %v %v", user.Id, user.Name)
		update, err := malpar.GetUserScoresByName(user.Name, 3, 3)
		if err != nil {
			mylog.Logger.Printf("Get score for %v: %v", user, err)
		} else {
			if len(update.AnimeList) > 0 || len(update.MangaList) > 0 {
				user.UpdateScoresFromList(update)
				_, err = s.db.Model(&user).Column("anime_scores", "manga_scores").Update()
				if err != nil {
					mylog.Logger.Printf("Update scores %v: %v", user.Name, err)
				}
				s.onUpdate(user)
			}

		}
	}
}
