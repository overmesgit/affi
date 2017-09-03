package main

import (
	"github.com/go-pg/pg"
	"sync"
	"updater"
)

func main() {
	db := pg.Connect(&pg.Options{
		User: "user", Password: "user", Database: "test",
	})
	//var err error
	//err = createSchema(db)
	//if err != nil {
	//	panic(err)
	//}

	updater.NewUserUpdater(db).Start()
	updater.NewScoreUpdater(db).Start(2)

	// Server because
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()

}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&updater.UserData{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
