package main

import (
	"flag"
	"fmt"
	"github.com/go-pg/pg"
	"server"
)

var (
	DBPassword     string
	DBName         string
	DBUser         string
	Host           string
	ScoresUpdaters int
)

func Empty(args ...string) bool {
	for _, val := range args {
		if val == "" {
			return true
		}
	}
	return false
}

func main() {
	flag.StringVar(&DBUser, "db_user", "", "postgresql db user")
	flag.StringVar(&DBPassword, "db_password", "", "postgresql db password")
	flag.StringVar(&DBName, "db_name", "", "postgresql db name")
	flag.StringVar(&Host, "host", "", "listening ip and port(127.0.0.1:8000)")
	flag.IntVar(&ScoresUpdaters, "scores_updaters", 3, "threads count for scores update")

	flag.Parse()
	if Empty(DBUser, DBPassword, DBName, Host) {
		fmt.Println(DBUser, DBPassword, DBName, Host)
		flag.PrintDefaults()
		panic("params not set")
	}

	db := pg.Connect(&pg.Options{
		User: DBUser, Password: DBPassword, Database: DBName,
	})

	myServer := server.NewServer(Host, db, ScoresUpdaters)
	myServer.Start()

}
