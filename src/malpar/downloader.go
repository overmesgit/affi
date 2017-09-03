package malpar

import (
	"fmt"
	"sync"
)

func fill_chain(start int, end int, c chan int) {
	i := start
	for {
		if i < end {
			c <- i
		} else {
			c <- 0
		}
		i++
	}
}

func get_user_scores(c chan int, res chan UserList, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	for id := range c {
		if id == 0 {
			break
		}
		userScores, err := GetUserScoresById(id, 3)
		if err == nil {
			res <- userScores
		} else {
			errors <- err
		}
		fmt.Println(id, userScores.UserName)
	}
}

func Download(start, end int, threads int, res chan UserList, errors chan error) {
	var wg sync.WaitGroup

	ch := make(chan int, 10*threads)

	go fill_chain(start, end, ch)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go get_user_scores(ch, res, errors, &wg)
	}

	wg.Wait()
}
