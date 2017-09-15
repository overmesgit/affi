package main

import (
	"github.com/go-pg/pg"
)

func main() {
	db := pg.Connect(&pg.Options{
		User: "user", Password: "user", Database: "test",
	})

	var err error
	err = createSchema(db)
	if err != nil {
		panic(err)
	}

	//offset := 0
	//limit := 1000
	//scores := 0
	//for {
	//	var users []updater.UserData
	//	err := db.Model(&users).Limit(limit).Offset(offset).Select()
	//	if len(users) == 0 {
	//		break
	//	}
	//	if err != nil {
	//		panic(err)
	//	} else {
	//		for i := range users {
	//			data := users[i]
	//			scores += len(data.AnimeScores) + len(data.MangaScores)
	//			animeArray := make(updater.UserScoreArray, 0)
	//			for s := range data.AnimeScores {
	//				score := data.AnimeScores[s]
	//				animeArray = append(animeArray, updater.NewUserScore(int(score.Id), int(score.Sc), int(score.St), int(score.Lu)))
	//			}
	//
	//			mangaArray := make(updater.UserScoreArray, 0)
	//			for s := range data.MangaScores {
	//				score := data.MangaScores[s]
	//				mangaArray = append(mangaArray, updater.NewUserScore(int(score.Id), int(score.Sc), int(score.St), int(score.Lu)))
	//			}
	//			newUserData := updater.UserScores{data.Id, data.Name, data.LastLogin, data.Gender,
	//				data.Birthday, data.Joined, data.Location, animeArray, mangaArray}
	//			err := db.Insert(&newUserData)
	//			if err != nil {
	//				fmt.Printf("error %v\n", err)
	//			}
	//
	//		}
	//	}
	//	offset += limit
	//	fmt.Printf("Loaded user before id %v\n", offset)
	//	fmt.Printf("Loaded scores %v\n", scores)
	//}

}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
