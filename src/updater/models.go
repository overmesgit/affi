package updater

import (
	"github.com/go-pg/pg"
	"time"
)

type LastUpdated struct {
	Id            int
	LastUpdatedId int
	LastProfileId int
}

func GetLastUpdate(db *pg.DB) (LastUpdated, error) {
	lastUpdate := LastUpdated{Id: 1, LastUpdatedId: 0}
	err := db.Select(&lastUpdate)
	if err == pg.ErrNoRows {
		err = db.Insert(&lastUpdate)
	}
	return lastUpdate, err
}

type UserData struct {
	Id         int
	Name       string
	LastOnline time.Time
	Gender     string
	Birthday   string
	Joined     string
	Location   string
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
