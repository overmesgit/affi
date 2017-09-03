package updater

import (
	"github.com/go-pg/pg"
	"malpar"
	"mylog"
)

type ScoreUpdater struct {
	db       *pg.DB
	toUpdate chan UserData
}

func NewScoreUpdater(db *pg.DB) *ScoreUpdater {
	return &ScoreUpdater{db: db, toUpdate: make(chan UserData, 1)}
}

func (s *ScoreUpdater) Start(threads int) {
	go s.fillUpdateQueue()
	for i := 0; i < threads; i++ {
		go s.startThread()
	}

}

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

func userListToScoresMap(list malpar.UserList) ([]UserScore, []UserScore) {
	animeResult := make([]UserScore, 0)

	for _, score := range list.AnimeList {
		animeResult = append(animeResult, UserScore{Id: score.Id, Sc: score.Score, St: score.Status, Lu: score.LastUpdate})
	}

	mangaResult := make([]UserScore, 0)
	for _, score := range list.MangaList {
		mangaResult = append(mangaResult, UserScore{Id: score.Id, Sc: score.Score, St: score.Status, Lu: score.LastUpdate})
	}

	return animeResult, mangaResult
}

func (s *ScoreUpdater) startThread() {
	for {
		user := <-s.toUpdate
		mylog.Logger.Printf("Get score for %v %v", user.Id, user.Name)
		update, err := malpar.GetUserScoresByName(user.Name, 3)
		if err != nil {
			mylog.Logger.Printf("Get score for %v: %v", user, err)
		} else {
			if len(update.AnimeList) > 0 || len(update.MangaList) > 0 {
				user.AnimeScores, user.MangaScores = userListToScoresMap(update)
				_, err = s.db.Model(&user).Column("anime_scores", "manga_scores").Update()
				if err != nil {
					mylog.Logger.Printf("Update scores %v: %v", user.Name, err)
				}
			}

		}
	}
}
