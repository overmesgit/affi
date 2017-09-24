package malpar

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type UserProfile struct {
	LastLogin time.Time
	Gender    string
	Birthday  string
	Joined    string
	Location  string
}

func GetUserNameById(userId int, retry int) (string, error) {
	fullUrl := mainUrl + fmt.Sprintf(commentsUrl, userId)
	var err error
	var resp *http.Response
	for i := 0; i <= retry; i++ {
		resp, err = http.Get(fullUrl)
		if err == nil {
			break
		}
		fmt.Println(err)

	}
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", ErrUserNotExist
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("User name not found %v", userId))
	}
	body := make([]byte, 250)
	resp.Body.Read(body)

	userNameRegexp := regexp.MustCompile(`<title>([\w\W]*)&#039;s Comments[\w\W]*<\/title>`)
	userNameMatch := userNameRegexp.FindStringSubmatch(string(body))
	if len(userNameMatch) == 0 {
		return "", errors.New(fmt.Sprintf("User name not found %v", userId))
	}
	return strings.TrimSpace(userNameMatch[1]), nil
}

func ParseLastOnline(strLastOnline string) (time.Time, error) {
	stringTimes := []string{"Now", "hour", "minutes", "Today", "Yesterday"}
	for _, strTime := range stringTimes {
		if strings.Contains(strLastOnline, strTime) {
			return time.Now(), nil
		}
	}

	if strings.Contains(strLastOnline, "Never") {
		return time.Date(2000, 0, 0, 0, 0, 0, 0, time.Local), nil
	}

	date, err := time.Parse("Jan 2, 2006 15:4 PM", strLastOnline)
	if err != nil {
		date, err = time.Parse("Jan 2, 15:4 PM", strLastOnline)
		if err != nil {
			fmt.Println(err)
			return time.Now(), err
		}
		date = date.AddDate(time.Now().Year(), 0, 0)
		return date, nil
	}
	return date, nil
}

func GetUserProfileDataByUserName(userName string, retry int) (UserProfile, error) {
	var userProfile = UserProfile{}
	fullUrl := mainUrl + fmt.Sprintf(profileUrl, userName)
	var err error
	var resp *http.Response
	for i := 0; i <= retry; i++ {
		resp, err = http.Get(fullUrl)
		if err == nil {
			break
		}
	}
	if err != nil {
		return userProfile, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return userProfile, errors.New(fmt.Sprintf("User not found %v", userName))
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return userProfile, err
	}

	lastOnline := regexp.MustCompile(`Last Online</span><(?:.*?)>(.*?)</span>`)
	lastOnlineMatch := lastOnline.FindStringSubmatch(string(body))
	lastLogin := time.Now()
	if len(lastOnlineMatch) > 0 {
		lastLogin, err = ParseLastOnline(lastOnlineMatch[1])
		if lastLogin.Before(time.Date(2000, 0, 0, 0, 0, 0, 0, time.Local)) {
			lastLogin = time.Date(2000, 0, 0, 0, 0, 0, 0, time.Local)
		}
		if err == nil {
			userProfile.LastLogin = lastLogin
		} else {
			fmt.Println(err)
		}
	}

	gender := regexp.MustCompile(`Gender</span><(?:.*?)>(.*?)</span>`)
	genderMatch := gender.FindStringSubmatch(string(body))
	if len(genderMatch) > 0 {
		userProfile.Gender = genderMatch[1]
	}
	birthday := regexp.MustCompile(`Birthday</span><(?:.*?)>(.*?)</span>`)
	birthdayMatch := birthday.FindStringSubmatch(string(body))
	if len(birthdayMatch) > 0 {
		userProfile.Birthday = birthdayMatch[1]
	}
	joined := regexp.MustCompile(`Joined</span><(?:.*?)>(.*?)</span>`)
	joinedMatch := joined.FindStringSubmatch(string(body))
	if len(joinedMatch) > 0 {
		userProfile.Joined = joinedMatch[1]
	}
	location := regexp.MustCompile(`Location</span><(?:.*?)>(.*?)</span>`)
	locationMatch := location.FindStringSubmatch(string(body))
	if len(locationMatch) > 0 {
		userProfile.Location = locationMatch[1]
	}

	return userProfile, nil
}
