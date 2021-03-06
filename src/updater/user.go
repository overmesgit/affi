package updater

import (
	"fmt"
	"github.com/go-pg/pg"
	"malpar"
	"mylog"
	"time"
)

const (
	CHECK_USER_RANGE = 100
)

type UserUpdater struct {
	db *pg.DB
}

func NewUserUpdater(db *pg.DB) *UserUpdater {
	return &UserUpdater{db: db}
}

func (u UserUpdater) Start() {
	go u.fullUpdater()
	go u.personalDataUpdate()
	go u.startNewUserUpdater()
}

func (u UserUpdater) startNewUserUpdater() {
	counter := 0
	for {
		u.getNewUsers()
		counter++
		time.Sleep(time.Hour)
	}
}

func (u UserUpdater) fullUpdater() {
	for {
		u.FullUpdate()
		update := LastUpdated{Id: 1, LastUpdatedId: 0}
		err := u.db.Update(&update)
		if err != nil {
			mylog.Logger.Printf("Update last update score: %v", err)
		}
		time.Sleep(30 * 24 * time.Hour)
	}
}

// Check all user ids, from 1 to last known
// Found deleted or missed users
func (u UserUpdater) FullUpdate() {
	lastUpdate, err := GetLastUpdate(u.db)
	if err != nil {
		mylog.Logger.Printf("Select last update id: %v", err)
	}

	lastUser, err := GetLastUser(u.db)
	lastId := uint(0)
	if err != nil {
		if err == pg.ErrNoRows {
			lastId = uint(100)
		} else {
			mylog.Logger.Printf("Get last user: %v", err)
		}
	} else {
		lastId = lastUser.Id
	}

	updateLogStack := 100
	updatedUsers := make([]string, 0, updateLogStack)
	for i := uint(lastUpdate.LastUpdatedId); i < lastId; i++ {
		username, err := malpar.GetUserNameById(int(i), 3)
		if err == malpar.ErrUserNotExist {
			err = u.db.Delete(&UserData{Id: i})
			if err == pg.ErrNoRows {
				continue
			}
			if err != nil {
				mylog.Logger.Printf("Delete user: %v", err)
			}
		}

		if err != nil {
			mylog.Logger.Printf("Get user data for %v: %v", i, err)
		} else {
			updatedUsers = append(updatedUsers, fmt.Sprintf("%v: %v", i, username))
			user := UserData{Id: i, Name: username}
			err = user.UpdateUserNameOrCreate(u.db)
			if err != nil {
				mylog.Logger.Printf("Update user %v: %v", username, err)
			} else {
				update := LastUpdated{Id: 1, LastUpdatedId: i}
				_, err := u.db.Model(&update).Column("last_updated_id").Update()
				if err != nil {
					mylog.Logger.Printf("Update last update score: %v", err)
				}
			}

		}
		if len(updatedUsers) >= updateLogStack {
			mylog.Logger.Printf("Updated users: %v", updatedUsers)
			updatedUsers = make([]string, 0, updateLogStack)
		}
	}
}

// Update users last login, birthday and so
func (u UserUpdater) personalDataUpdate() {
	lastUpdate, err := GetLastUpdate(u.db)
	if err != nil {
		mylog.Logger.Printf("Select last update id: %v", err)
	}

	for {
		var nextUser UserData
		err := u.db.Model(&nextUser).Where("id > ?", lastUpdate.LastProfileId).Order("id asc").First()
		if err == pg.ErrNoRows {
			mylog.Logger.Printf("User profiles updated, restart, last id: %v", lastUpdate.LastProfileId)
			lastUpdate.LastProfileId = 0
			_, err = u.db.Model(&lastUpdate).Column("last_profile_id").Update()
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
			lastUpdate.LastProfileId++
			continue
		}

		userData := UserData{Id: nextUser.Id, Name: nextUser.Name, LastLogin: profile.LastLogin,
			Gender: profile.Gender, Birthday: profile.Birthday, Joined: profile.Joined, Location: profile.Location}
		_, err = u.db.Model(&userData).Column("name", "last_login", "gender", "birthday", "joined", "location").Update()
		if err != nil {
			mylog.Logger.Printf("Update user profile %v: %v", nextUser.Name, err)
		}

		lastUpdate.LastProfileId = nextUser.Id
		_, err = u.db.Model(&lastUpdate).Column("last_profile_id").Update()
		if err != nil {
			mylog.Logger.Printf("Update LastUpdate: %v", err)
		}
	}
}

// Get last known user and check user existence in neighbour range
// Commonly last user will be last signup user
func (u UserUpdater) getNewUsers() {
	lastUser, err := GetLastUser(u.db)
	if err != nil {
		mylog.Logger.Printf("Get last user: %v", err)
	}

	firstId := lastUser.Id - CHECK_USER_RANGE
	if firstId < 1 {
		firstId = 1
	}

	for userId := firstId; userId < lastUser.Id+CHECK_USER_RANGE; userId++ {
		username, err := malpar.GetUserNameById(int(userId), 3)
		if err != nil {
			mylog.Logger.Printf("User update: %v %v", userId, err)
		} else {
			mylog.Logger.Printf("Updating user: %v %v", userId, username)
			user := UserData{Id: userId, Name: username}
			err := user.UpdateUserNameOrCreate(u.db)
			if err != nil {
				mylog.Logger.Printf("Update user data: %v %v %v", user.Id, user.Name, err)
			}

		}
	}
}
