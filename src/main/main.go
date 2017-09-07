package main

import (
	"github.com/go-pg/pg"
	"server"
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

	myServer := server.NewServer("127.0.0.1:8000", db)
	myServer.Start()

}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&server.Result{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
