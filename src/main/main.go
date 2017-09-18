package main

import (
	"github.com/go-pg/pg"
	"server"
)

func main() {
	db := pg.Connect(&pg.Options{
		User: "user", Password: "user", Database: "test",
	})

	myServer := server.NewServer("0.0.0.0:8000", db)
	myServer.Start()

}
