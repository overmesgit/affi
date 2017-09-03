package main

import (
	"github.com/go-pg/pg"
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

	updater.StartUserUpdater(db)

}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&updater.LastUpdated{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
