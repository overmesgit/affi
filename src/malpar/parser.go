package malpar

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
)

const (
	mainUrl     = "http://myanimelist.net"
	commentsUrl = "/comments.php?id=%v"
	profileUrl  = "/profile/%v"
	apiUrl      = "/malappinfo.php?u=%v&status=all&type=%s"
)

var (
	ErrTooMuchRequests = Errorf("too much requests")
	ErrUserNotExist    = Errorf("user not exist")
)

type Error struct {
	s string
}

func (err Error) Error() string {
	return err.s
}

func Errorf(s string, args ...interface{}) Error {
	return Error{s: fmt.Sprintf(s, args...)}
}

type ParserError struct {
	Msg  string
	Name string
	Id   int
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("id: %v name: %v msg: %v", e.Name, e.Id, e.Msg)
}

type UserListParser interface {
	ParseAnime(io.Reader)
	ParseManga(io.Reader)
	ScoresCount() int
	ToArrayFormat()
}

func TitlesAvg(scoresSlice interface{}) float32 {
	sum := 0
	count := 0
	scores := reflect.ValueOf(scoresSlice)
	for i := 0; i < scores.Len(); i++ {
		s := scores.Index(i).Interface().(TitleFace)
		if s.GetScore() > 0 {
			sum += int(s.GetScore())
			count++
		}
	}
	return float32(sum) / float32(count)
}

type UserList struct {
	UserId    int
	UserName  string
	AnimeList []AnimeTitle
	MangaList []MangaTitle
}

func (l *UserList) AnimeAvg() float32 {
	return TitlesAvg(l.AnimeList)
}

func (l *UserList) MangaAvg() float32 {
	return TitlesAvg(l.MangaList)
}

func (l *UserList) ToArrayFormat() map[string][3]uint {
	result := map[string][3]uint{}
	for i := range l.AnimeList {
		current := l.AnimeList[i]
		scoreArray := [3]uint{current.Score, current.Status, uint(current.LastUpdate)}
		result[strconv.Itoa(int(current.Id))] = scoreArray
	}
	for i := range l.MangaList {
		current := l.MangaList[i]
		scoreArray := [3]uint{current.Score, current.Status, uint(current.LastUpdate)}
		result[strconv.Itoa(-int(current.Id))] = scoreArray
	}
	return result
}

func (l *UserList) ScoresCount() int {
	return len(l.AnimeList) + len(l.MangaList)
}

func (l *UserList) ParseTitles(data []byte, type_ string) error {
	var err error

	switch {
	case type_ == "anime":
		v := AnimeResult{}
		err = xml.Unmarshal(data, &v)
		if err != nil {
			return err
		}

		for i := range v.TitleList {
			v.TitleList[i].SetLastUpdate()
		}
		l.AnimeList = v.TitleList
		l.UserId = v.MyInfo.UserId
	case type_ == "manga":
		v := MangaResult{}
		err = xml.Unmarshal(data, &v)
		if err != nil {
			fmt.Printf("error: %v", err)
			return err
		}

		for i := range v.TitleList {
			v.TitleList[i].SetLastUpdate()
		}
		l.MangaList = v.TitleList
		l.UserId = v.MyInfo.UserId
	case true:
		return errors.New("Wrong type")
	}
	return nil

}

func getUserApiPage(userName string, content string) ([]byte, error) {
	url := mainUrl + fmt.Sprintf(apiUrl, userName, content)

	resp, err := http.Get(url)

	var body []byte
	if err != nil {
		return body, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 429 {
			return body, ErrTooMuchRequests
		}
		return body, errors.New(fmt.Sprintf("User page error %v %v", url, resp.StatusCode))
	}
	return ioutil.ReadAll(resp.Body)

	v := MalApiError{}
	err = xml.Unmarshal(body, &v)
	if v.Error != "" {
		return body, errors.New(v.Error)
	}
	if v.MyInfo.UserId == 0 {
		return body, errors.New(fmt.Sprintf("User '%v' not found", userName))
	}
	return body, nil

}
