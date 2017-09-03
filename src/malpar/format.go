package malpar

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type Title struct {
	Status      uint   `xml:"my_status"`
	Score       uint   `xml:"my_score"`
	LastUpdate  uint   `xml:"my_last_updated"`
	MyStartDate string `xml:"my_start_date"`
	MyLastDate  string `xml:"my_finish_date"`
}

type TitleFace interface {
	GetScore() uint
}

type AnimeTitle struct {
	Title
	Id uint `xml:"series_animedb_id"`
}

func (t AnimeTitle) GetScore() uint {
	return t.Score
}

type MangaTitle struct {
	Title
	Id uint `xml:"series_mangadb_id"`
}

func (t MangaTitle) GetScore() uint {
	return t.Score
}

type MyInfo struct {
	UserId   int    `xml:"user_id"`
	Username string `xml:"user_name"`
}

type Result struct {
	XMLName xml.Name `xml:"myanimelist"`
	MyInfo  MyInfo   `xml:"myinfo"`
}

type MalApiError struct {
	XMLName xml.Name `xml:"myanimelist"`
	Error   string   `xml:"error"`
	MyInfo  MyInfo   `xml:"myinfo"`
}

type AnimeResult struct {
	Result
	TitleList []AnimeTitle `xml:"anime"`
}

type MangaResult struct {
	Result
	TitleList []MangaTitle `xml:"manga"`
}

type TitleLastUpdater interface {
	SetLastUpdate()
}

func GetMALDate(date string) uint {
	layout := "2006-01-02"
	date = strings.Replace(date, "-00", "-01", -1)
	myLast, err := time.Parse(layout, date)
	//int32 unix time
	if myLast.Year() < 1950 || myLast.Year() > 2030 {
		return 0
	}
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return uint(myLast.Unix())
}

func (t *Title) SetLastUpdate() {
	if t.LastUpdate == 0 {
		if !strings.Contains(t.MyLastDate, "0000") {
			t.LastUpdate = GetMALDate(t.MyLastDate)
		} else {
			if !strings.Contains(t.MyStartDate, "0000") {
				t.LastUpdate = GetMALDate(t.MyStartDate)
			}
		}
	}
}
