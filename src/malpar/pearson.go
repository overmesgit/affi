package malpar

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type UserScore struct {
	Id    int32
	Score int8
}

type UserScoresSlice []UserScore

func (a UserScoresSlice) Len() int           { return len(a) }
func (a UserScoresSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a UserScoresSlice) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a UserScoresSlice) Sort()              { sort.Sort(a) }
func (a UserScoresSlice) Search(id int32) (int8, bool) {
	res := sort.Search(len(a), func(i int) bool {
		return a[i].Id >= id
	})
	if res == len(a) || a[res].Id != id {
		return -1.0, false
	} else {
		return a[res].Score, true
	}
}

func ArraysToPearson(userScoresA, userScoresB UserScoresSlice) (int32, float32) {
	// Так как сложность O(m*ln(n))
	if len(userScoresA) > len(userScoresB) {
		userScoresA, userScoresB = userScoresB, userScoresA
	}
	numerator := float32(0.0)
	denominatorA := float32(0.0)
	denominatorB := float32(0.0)
	shared := int32(0)

	sharedScores := make([][2]int8, 0)
	scoresASum := 0
	scoresBSum := 0
	for i := range userScoresA {
		item, scoreA := userScoresA[i].Id, userScoresA[i].Score
		scoreB, ok := userScoresB.Search(item)
		if ok {
			sharedScores = append(sharedScores, [2]int8{scoreA, scoreB})
			scoresASum += int(scoreA)
			scoresBSum += int(scoreB)
			shared++
		}
	}

	if shared == 0 {
		return shared, float32(0.0)
	}
	aAvg := float32(scoresASum) / float32(shared)
	bAvg := float32(scoresBSum) / float32(shared)

	for i := range sharedScores {
		scoreA, scoreB := float32(sharedScores[i][0])-aAvg, float32(sharedScores[i][1])-bAvg
		numerator += scoreA * scoreB
		denominatorA += scoreA * scoreA
		denominatorB += scoreB * scoreB
	}
	if denominatorA*denominatorB == float32(0.0) {
		return shared, float32(0.0)
	}
	return shared, float32(float64(numerator) / (math.Sqrt(float64(denominatorA * denominatorB))))
}

type PearsonResult struct {
	Id      int32
	Shared  int32
	Pearson float32
}

type PearsonSlice []PearsonResult

func (a PearsonSlice) Len() int      { return len(a) }
func (a PearsonSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PearsonSlice) Less(i, j int) bool {
	return a[i].Pearson < a[j].Pearson
}

type Pearson struct {
	Scores       []UserScoresSlice
	IndexReplace map[int32]int32
	IdReplace    map[int32]int32
}

func NewPearson() *Pearson {
	return &Pearson{make([]UserScoresSlice, 0), make(map[int32]int32, 100), make(map[int32]int32, 100)}
}

func (p *Pearson) UpdateUserSlices(userId int, scores UserScoresSlice) {
	if userId != 0 && len(scores) > 0 {
		scores.Sort()

		id := int32(userId)
		currentIndex, ok := p.IdReplace[id]
		if ok {
			p.Scores[currentIndex] = scores
		} else {
			index := int32(len(p.Scores))
			p.IndexReplace[index] = id
			p.IdReplace[id] = index
			p.Scores = append(p.Scores, scores)
		}
	}
}

func (p *Pearson) UpdateUserList(list UserList) {
	userSlice := p.UserListToScores(list, true, true)
	p.UpdateUserSlices(list.UserId, userSlice)
}

func (p *Pearson) UserListToScores(list UserList, anime, manga bool) UserScoresSlice {
	res := make(UserScoresSlice, 0)
	if anime {
		for j := range list.AnimeList {
			if list.AnimeList[j].Score > 0 {
				score := UserScore{int32(list.AnimeList[j].Id), int8(list.AnimeList[j].Score)}
				res = append(res, score)
			}
		}
	}
	if manga {
		for j := range list.MangaList {
			if list.MangaList[j].Score > 0 {
				score := UserScore{-int32(list.MangaList[j].Id), int8(list.MangaList[j].Score)}
				res = append(res, score)
			}
		}
	}
	res.Sort()
	return res
}

func (p *Pearson) Count(list UserList, minShare int, anime, manga bool, update bool) PearsonSlice {
	userSlice := p.UserListToScores(list, anime, manga)
	if update {
		p.UpdateUserSlices(list.UserId, userSlice)
	}

	start := time.Now().UnixNano()

	pearsonResults := make(PearsonSlice, 0)
	for i := 0; i < len(p.Scores); i++ {
		shared, pearson := ArraysToPearson(userSlice, p.Scores[i])
		sharedUserId := p.IndexReplace[int32(i)]
		if int(shared) > minShare && sharedUserId != int32(list.UserId) {
			pearsonResults = append(pearsonResults, PearsonResult{sharedUserId, shared, pearson})
		}
	}
	sort.Sort(sort.Reverse(pearsonResults))

	end := float64(time.Now().UnixNano()-start) / float64(time.Millisecond)
	fmt.Printf("pearson for id %v name %v counted %v scores %v\n", list.UserId, list.UserName, end, len(p.Scores))

	return pearsonResults
}
