package server

import (
	"counter"
	"time"
)

const (
	COMPLITED = "complited"
	COUNTING  = "counting"
	PENDING   = "pending"
	ERROR     = "error"
)

type Result struct {
	Id        int
	UserName  string
	Status    string
	Pearson   counter.PearsonSlice
	Compare   int
	Error     string
	Created   time.Time
	Completed time.Time
	Manga     bool
	Anime     bool
	Share     int
	Progress  int
}

func NewResult() *Result {
	res := Result{Pearson: make(counter.PearsonSlice, 0), Created: time.Now(), Status: PENDING}
	return &res
}
