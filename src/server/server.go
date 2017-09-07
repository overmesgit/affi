package server

import (
	"counter"
	"github.com/go-pg/pg"
	"html/template"
	"mylog"
	"net/http"
	//"updater"
	"updater"
)

var homeTempl *template.Template

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	// TODO: remove
	//homeTempl = template.Must(template.ParseFiles("templates/home.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

type BaseJsonResponse struct {
	Status  string
	Message string
}

type DataJsonResponse struct {
	BaseJsonResponse
	Data interface{}
}

func GetErrorResponse(err error) interface{} {
	return BaseJsonResponse{"error", err.Error()}
}

func GetDataResponse(data interface{}) interface{} {
	return DataJsonResponse{BaseJsonResponse{"ok", ""}, data}
}

func FormatResponse(data interface{}, error error) interface{} {
	if error != nil {
		return GetErrorResponse(error)
	} else {
		return GetDataResponse(data)
	}
}

type Server struct {
	Host    string
	counter *counter.PearsonCounter
	db      *pg.DB
}

func NewServer(host string, db *pg.DB) Server {
	return Server{host, counter.NewPearsonCounter(db), db}
}

func (c *Server) Start() {
	homeTempl = template.Must(template.ParseFiles("templates/home.html"))

	c.counter.Prepare()

	updater.NewUserUpdater(c.db).Start()

	updater.NewScoreUpdater(c.db, func(user updater.UserData) {
		c.counter.UpdateChan <- user
	}).Start(2)

	c.counter.Start()

	mylog.Logger.Printf("start %v", c.Host)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/api/result/", c.serveResult)
	http.HandleFunc("/", serveHome)

	err := http.ListenAndServe(c.Host, nil)
	if err != nil {
		panic(err)
	}

	mylog.Logger.Println("stop")
}
