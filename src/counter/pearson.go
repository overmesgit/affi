package counter

import (
	"math"
	"mylog"
	"sort"
	"time"
	"updater"
)

type Pearson struct {
	AnimeIndexes [][]int16
	AnimeScores  [][]int8
	MangaIndexes [][]int16
	MangaScores  [][]int8
	// UserIndex -> UserId
	UserIndexReplace map[int32]int32
	// UserId -> UserIndex
	UserIdReplace map[int32]int32
	// AnimeId -> AnimeIndex
	AnimeIdReplace map[uint]int16
	// MangaId -> MangaIndex
	MangaIdReplace map[uint]int16
}

func NewPearson() *Pearson {
	return &Pearson{AnimeIndexes: make([][]int16, 0), AnimeScores: make([][]int8, 0),
		MangaIndexes: make([][]int16, 0), MangaScores: make([][]int8, 0),
		UserIndexReplace: make(map[int32]int32, 100), UserIdReplace: make(map[int32]int32, 100),
		AnimeIdReplace: make(map[uint]int16), MangaIdReplace: make(map[uint]int16)}
}

func getIdReplace(id uint, replace map[uint]int16) int16 {
	index, ok := replace[id]
	if !ok {
		index = int16(len(replace))
		replace[id] = index

	}
	return index
}

func (p *Pearson) getUserIndex(userId int) int32 {
	index, ok := p.UserIdReplace[int32(userId)]
	if !ok {
		index = int32(len(p.UserIdReplace))
		p.UserIdReplace[int32(userId)] = index
		p.UserIndexReplace[index] = int32(userId)

	}
	return index
}

func (p *Pearson) getSlice(userList []updater.UserScore, replace map[uint]int16) ([]int16, []int8) {
	indexes := make([]int16, 0)
	scores := make([]int8, 0)
	for i := range userList {
		if userList[i].Sc > 0 {
			itemIndex := getIdReplace(userList[i].Id, replace)
			indexes = append(indexes, itemIndex)
			scores = append(scores, int8(userList[i].Sc))
		}
	}
	return indexes, scores
}

func (p *Pearson) UpdateUserSlices(user updater.UserData) int32 {
	userIndex := int32(0)
	if len(user.AnimeScores) > 0 || len(user.MangaScores) > 0 {
		sort.Slice(user.AnimeScores, func(i, j int) bool {
			return user.AnimeScores[i].Id < user.AnimeScores[j].Id
		})

		animeIndexes, animeScores := p.getSlice(user.AnimeScores, p.AnimeIdReplace)
		mangaIndexes, mangaScores := p.getSlice(user.MangaScores, p.MangaIdReplace)

		userIndex = p.getUserIndex(user.Id)
		if userIndex >= int32(len(p.AnimeScores)) {
			p.AnimeIndexes = append(p.AnimeIndexes, animeIndexes)
			p.AnimeScores = append(p.AnimeScores, animeScores)

			p.MangaIndexes = append(p.MangaIndexes, mangaIndexes)
			p.MangaScores = append(p.MangaScores, mangaScores)
		} else {
			p.AnimeIndexes[userIndex] = animeIndexes
			p.AnimeScores[userIndex] = animeScores

			p.MangaIndexes[userIndex] = mangaIndexes
			p.MangaScores[userIndex] = mangaScores
		}
	}
	return userIndex
}

type PearsonResult struct {
	Id      int32
	Shared  int
	Pearson float32
}

type PearsonSlice []PearsonResult

func (a PearsonSlice) Len() int      { return len(a) }
func (a PearsonSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PearsonSlice) Less(i, j int) bool {
	return a[i].Pearson < a[j].Pearson
}

func (p *Pearson) Count(user updater.UserData, minShare int, anime, manga bool) PearsonSlice {
	userIndex := p.UpdateUserSlices(user)

	start := time.Now().UnixNano()

	pearsonHeap := make(PearsonResultHeap, 0)
	for otherIndex := 0; otherIndex < len(p.AnimeIndexes); otherIndex++ {
		otherInt32 := int32(otherIndex)
		if userIndex != otherInt32 {
			shared, pearson := p.IndexesToPearson(userIndex, otherInt32, anime, manga)
			sharedUserId := p.UserIndexReplace[otherInt32]
			if int(shared) > minShare {
				pearsonHeap.Push(PearsonResult{sharedUserId, shared, pearson})
				if len(pearsonHeap) > 100 {
					pearsonHeap.Pop()
				}
			}
		}

	}

	res := PearsonSlice(pearsonHeap)
	sort.Sort(sort.Reverse(res))

	end := float64(time.Now().UnixNano()-start) / float64(time.Millisecond)
	mylog.Logger.Printf("Pearson for %v: %v counted in %v ms Scores: %v\n", user.Id, user.Name, end, len(p.AnimeIndexes))

	return res
}

func scoresSum(indexesA, indexesB []int16, scoresA, scoresB []int8) (int, int, [][2]int8) {
	scoresASum := 0
	scoresBSum := 0
	sharedScores := make([][2]int8, 0)
	for a, b := 0, 0; a < len(indexesA) && b < len(indexesB); {
		if indexesA[a] == indexesB[b] {
			scoresASum += int(scoresA[a])
			scoresBSum += int(scoresB[b])
			sharedScores = append(sharedScores, [2]int8{scoresA[a], scoresB[b]})

			a++
			b++
		} else {
			if indexesA[a] < indexesB[b] {
				a++
			} else {
				b++
			}
		}
	}
	return scoresASum, scoresBSum, sharedScores
}

func (p *Pearson) IndexesToPearson(indexA, indexB int32, anime, manga bool) (int, float32) {
	numerator := float32(0.0)
	denominatorA := float32(0.0)
	denominatorB := float32(0.0)

	sharedScores := make([][2]int8, 0)
	scoresASum := 0
	scoresBSum := 0

	if anime {
		aSum, bSub, scores := scoresSum(p.AnimeIndexes[indexA], p.AnimeIndexes[indexB], p.AnimeScores[indexA], p.AnimeScores[indexB])
		scoresASum += aSum
		scoresBSum += bSub
		sharedScores = append(sharedScores, scores...)
	}

	if manga {
		aSum, bSub, scores := scoresSum(p.MangaIndexes[indexA], p.MangaIndexes[indexB], p.MangaScores[indexA], p.MangaScores[indexB])
		scoresASum += aSum
		scoresBSum += bSub
		sharedScores = append(sharedScores, scores...)
	}

	if len(sharedScores) == 0 {
		return 0, float32(0.0)
	}
	aAvg := float32(scoresASum) / float32(len(sharedScores))
	bAvg := float32(scoresBSum) / float32(len(sharedScores))

	for i := range sharedScores {
		scoreA, scoreB := float32(sharedScores[i][0])-aAvg, float32(sharedScores[i][1])-bAvg
		numerator += scoreA * scoreB
		denominatorA += scoreA * scoreA
		denominatorB += scoreB * scoreB
	}
	if denominatorA*denominatorB == float32(0.0) {
		return len(sharedScores), float32(0.0)
	}
	return len(sharedScores), float32(float64(numerator) / (math.Sqrt(float64(denominatorA * denominatorB))))
}
