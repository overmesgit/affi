package malpar

import "time"

func GetUserScoresById(userId int, retry int) (UserList, error) {
	userName, err := GetUserNameById(userId, 3)
	if err != nil {
		return UserList{UserId: userId}, err
	}
	userList, err := GetUserScoresByName(userName, 3, 3)
	userList.UserId = userId
	return userList, err
}

func GetUserScoresByName(userName string, retry int, sleep time.Duration) (UserList, error) {
	userList := UserList{UserName: userName}
	for _, content := range [2]string{"anime", "manga"} {
		var err error
		var body []byte
		for try := 1; try == 1 || (try <= retry && err != nil); try++ {
			body, err = getUserApiPage(userName, content)
			time.Sleep(sleep * time.Second)
		}
		if err != nil {
			return userList, err
		}
		err = userList.ParseTitles(body, content)
		if err != nil {
			return userList, err
		}
	}
	return userList, nil
}
