package updater

import (
	"github.com/go-pg/pg"
	"malpar"
	"time"
)

type LastUpdated struct {
	Id            int
	LastUpdatedId uint
	LastProfileId uint
	LastScoreId   uint
}

func GetLastUpdate(db *pg.DB) (LastUpdated, error) {
	lastUpdate := LastUpdated{Id: 1, LastUpdatedId: 0}
	err := db.Select(&lastUpdate)
	if err == pg.ErrNoRows {
		err = db.Insert(&lastUpdate)
	}
	return lastUpdate, err
}

// Let's keep posgresql space with short json keys
type UserScore struct {
	Id uint  // ItemId
	Sc uint8 // Score
	Lu uint  // LastUpdate
	St uint8 // Status
}

type UserData struct {
	Id          uint
	Name        string
	LastLogin   time.Time
	Gender      string
	Birthday    string
	Joined      string
	Location    string
	AnimeScores []UserScore
	MangaScores []UserScore
}

func GetLastUser(db *pg.DB) (UserData, error) {
	var lastUser UserData
	err := db.Model(&lastUser).Order("id desc").First()
	return lastUser, err
}

func (u UserData) UpdateUserNameOrCreate(db *pg.DB) error {
	err := db.Select(&u)
	if err == pg.ErrNoRows {
		return db.Insert(&u)
	} else {
		_, err = db.Model(&u).Column("name").Update()
		return err
	}
}

func (u *UserData) UpdateScoresFromList(list malpar.UserList) {
	animeResult := make([]UserScore, 0)
	for _, score := range list.AnimeList {
		animeResult = append(animeResult, UserScore{Id: score.Id, Sc: uint8(score.Score), St: uint8(score.Status), Lu: score.LastUpdate})
	}
	u.AnimeScores = animeResult

	mangaResult := make([]UserScore, 0)
	for _, score := range list.MangaList {
		mangaResult = append(mangaResult, UserScore{Id: score.Id, Sc: uint8(score.Score), St: uint8(score.Status), Lu: score.LastUpdate})
	}
	u.MangaScores = mangaResult
}
