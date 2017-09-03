package updater

import (
	"github.com/go-pg/pg"
	"malpar"
	"mylog"
	"time"
)

const (
	CHECK_USER_RANGE = 100
)

var db *pg.DB

func StartUserUpdater(useDB *pg.DB) {
	db = useDB
	go fullUpdater()
	go personalDataUpdate()
	counter := 0
	for {
		go getNewUsers()
		counter++
		time.Sleep(time.Hour)
	}
}

func fullUpdater() {
	for {
		FullUpdate()
		update := LastUpdated{Id: 1, LastUpdatedId: 0}
		err := db.Update(&update)
		if err != nil {
			mylog.Logger.Printf("Update last update score: %v", err)
		}
		time.Sleep(30 * 24 * time.Hour)
	}
}

// Check all user ids, from 1 to last known
// Found deleted or missed users
func FullUpdate() {
	lastUpdate, err := GetLastUpdate(db)
	if err != nil {
		mylog.Logger.Printf("Select last update id: %v", err)
	}

	lastUser, err := GetLastUser(db)
	if err != nil {
		mylog.Logger.Printf("Get last user: %v", err)
	}

	for i := lastUpdate.LastUpdatedId; i < lastUser.Id; i++ {
		username, err := malpar.GetUserNameById(i, 3)
		if err == pg.ErrNoRows {
			err = db.Delete(&UserData{Id: i})
			if err != nil {
				mylog.Logger.Printf("Delete user: %v", err)
			}
		}

		if err != nil {
			mylog.Logger.Printf("Get user data: %v", err)
		} else {
			mylog.Logger.Printf("Get user data: %v %v", i, username)
			user := UserData{Id: i, Name: username}
			err = user.UpdateUserNameOrCreate(db)
			if err != nil {
				mylog.Logger.Printf("Update user %v: %v", username, err)
			} else {
				update := LastUpdated{Id: 1, LastUpdatedId: i}
				_, err := db.Model(&update).Column("last_updated_id").Update()
				if err != nil {
					mylog.Logger.Printf("Update last update score: %v", err)
				}
			}

		}
	}
}

// Update users last login, birthday and so
func personalDataUpdate() {
	lastUpdate, err := GetLastUpdate(db)
	if err != nil {
		mylog.Logger.Printf("Select last update id: %v", err)
	}

	for {
		var nextUser UserData
		err := db.Model(&nextUser).Where("id > ?", lastUpdate.LastProfileId).Order("id asc").First()
		if err == pg.ErrNoRows {
			mylog.Logger.Printf("User profiles updated, restart, last id: %v", lastUpdate.LastProfileId)
			lastUpdate.LastProfileId = 0
			_, err = db.Model(&lastUpdate).Column("last_profile_id").Update()
			if err != nil {
				mylog.Logger.Printf("Update LastUpdate: %v", err)
			}
			continue
		}

		if err != nil {
			mylog.Logger.Printf("Get next user: %v", err)
			continue
		}

		profile, err := malpar.GetUserProfileDataByUserName(nextUser.Name, 3)
		if err != nil {
			mylog.Logger.Printf("Get user profile %v: %v", nextUser.Name, err)
			continue
		}

		userData := UserData{nextUser.Id, nextUser.Name, profile.LastLogin, profile.Gender, profile.Birthday, profile.Joined, profile.Location}
		err = db.Update(&userData)
		if err != nil {
			mylog.Logger.Printf("Update user profile %v: %v", nextUser.Name, err)
		}

		lastUpdate.LastProfileId = nextUser.Id
		_, err = db.Model(&lastUpdate).Column("last_profile_id").Update()
		if err != nil {
			mylog.Logger.Printf("Update LastUpdate: %v", err)
		}
	}
}

// Get last known user and check user existence in neighbour range
// Commonly last user will be last signup user
func getNewUsers() {
	lastUser, err := GetLastUser(db)
	if err != nil {
		mylog.Logger.Printf("Get last user: %v", err)
	}

	firstId := lastUser.Id - CHECK_USER_RANGE
	if firstId < 1 {
		firstId = 1
	}

	for userId := firstId; userId < lastUser.Id+CHECK_USER_RANGE; userId++ {
		username, err := malpar.GetUserNameById(userId, 3)
		if err != nil {
			mylog.Logger.Printf("User update: %v %v", userId, err)
		} else {
			mylog.Logger.Printf("Updating user: %v %v", userId, username)
			user := UserData{Id: userId, Name: username}
			err := user.UpdateUserNameOrCreate(db)
			if err != nil {
				mylog.Logger.Printf("Update user data: %v %v %v", user.Id, user.Name, err)
			}

		}
	}
}
