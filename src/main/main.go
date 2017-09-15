package main

import (
	"github.com/go-pg/pg"
	"server"
)

func main() {
	db := pg.Connect(&pg.Options{
		User: "user", Password: "user", Database: "test",
	})

	myServer := server.NewServer("127.0.0.1:8000", db)
	myServer.Start()

}
