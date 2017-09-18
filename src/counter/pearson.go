package counter

import (
	"container/heap"
	"math"
	"mylog"
	"sort"
	"time"
	"updater"
)

type Pearson struct {
	AnimeIndexes [][]uint16
	AnimeScores  [][]uint8
	MangaIndexes [][]uint16
	MangaScores  [][]uint8
	// UserIndex -> UserId
	UserIndexToId map[uint32]uint32
	// UserId -> UserIndex
	UserIdToIndex map[uint32]uint32
	// AnimeId -> AnimeIndex
	AnimeIdToIndex map[uint]uint16
	// MangaId -> MangaIndex
	MangaIdToIndex map[uint]uint16
}

func NewPearson() *Pearson {
	return &Pearson{AnimeIndexes: make([][]uint16, 0), AnimeScores: make([][]uint8, 0),
		MangaIndexes: make([][]uint16, 0), MangaScores: make([][]uint8, 0),
		UserIndexToId: make(map[uint32]uint32, 100), UserIdToIndex: make(map[uint32]uint32, 100),
		AnimeIdToIndex: make(map[uint]uint16), MangaIdToIndex: make(map[uint]uint16)}
}

func getIdReplace(id uint, replace map[uint]uint16) uint16 {
	index, ok := replace[id]
	if !ok {
		index = uint16(len(replace))
		replace[id] = index

	}
	return index
}

func (p *Pearson) getUserIndex(userId uint) uint32 {
	index, ok := p.UserIdToIndex[uint32(userId)]
	if !ok {
		index = uint32(len(p.UserIdToIndex))
		p.UserIdToIndex[uint32(userId)] = index
		p.UserIndexToId[index] = uint32(userId)

	}
	return index
}

func (p *Pearson) getSlice(userList []updater.UserScore, replace map[uint]uint16) ([]uint16, []uint8) {
	sort.Slice(userList, func(i, j int) bool {
		return getIdReplace(userList[i].Id, replace) < getIdReplace(userList[j].Id, replace)
	})
	indexes := make([]uint16, 0)
	scores := make([]uint8, 0)
	// first 4 bits for first score, second 4 bits for second
	scoresIndex := 0
	for i := range userList {
		if userList[i].Sc > 0 {
			itemIndex := getIdReplace(userList[i].Id, replace)
			indexes = append(indexes, itemIndex)

			if scoresIndex%2 == 0 {
				scores = append(scores, userList[i].Sc<<4)
			} else {
				scores[scoresIndex/2] += userList[i].Sc
			}
			scoresIndex++
		}
	}
	return indexes, scores
}

func (p *Pearson) UpdateUserSlices(user updater.UserData) uint32 {
	userIndex := uint32(0)
	if len(user.AnimeScores) > 0 || len(user.MangaScores) > 0 {
		animeIndexes, animeScores := p.getSlice(user.AnimeScores, p.AnimeIdToIndex)
		mangaIndexes, mangaScores := p.getSlice(user.MangaScores, p.MangaIdToIndex)

		userIndex = p.getUserIndex(user.Id)
		if userIndex >= uint32(len(p.AnimeScores)) {
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
	Id      uint32
	Shared  int
	Pearson float32
}

type PearsonSlice []PearsonResult

func (a PearsonSlice) Len() int      { return len(a) }
func (a PearsonSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PearsonSlice) Less(i, j int) bool {
	return a[i].Pearson < a[j].Pearson
}

func (p *Pearson) Count(user updater.UserData, minShare int, anime, manga bool) (PearsonSlice, int) {
	userIndex := p.UpdateUserSlices(user)

	start := time.Now().UnixNano()

	compare := 0
	pearsonHeap := &PearsonResultHeap{}
	for otherIndex := 0; otherIndex < len(p.AnimeIndexes); otherIndex++ {
		otherInt32 := uint32(otherIndex)
		if userIndex != otherInt32 {
			shared, pearson := p.IndexesToPearson(userIndex, otherInt32, anime, manga)
			sharedUserId := p.UserIndexToId[otherInt32]
			if int(shared) > minShare {
				compare++
				heap.Push(pearsonHeap, PearsonResult{sharedUserId, shared, pearson})
				if pearsonHeap.Len() > 100 {
					heap.Pop(pearsonHeap)
				}
			}
		}

	}

	res := PearsonSlice(*pearsonHeap)
	sort.Sort(sort.Reverse(res))

	end := float64(time.Now().UnixNano()-start) / float64(time.Millisecond)
	mylog.Logger.Printf("Pearson for %v: %v counted in %v ms Scores: %v\n", user.Id, user.Name, end, len(p.AnimeIndexes))

	return res, compare
}

func getScore(scores []uint8, i int) uint8 {
	if i%2 == 0 {
		return scores[i/2] >> 4
	} else {
		return scores[i/2] << 4 >> 4
	}
}

func scoresSum(indexesA, indexesB []uint16, scoresA, scoresB []uint8) (int, int, [][2]uint8) {
	scoresASum := 0
	scoresBSum := 0
	sharedScores := make([][2]uint8, 0)
	for a, b := 0, 0; a < len(indexesA) && b < len(indexesB); {
		if indexesA[a] == indexesB[b] {
			scoreA := getScore(scoresA, a)
			scoreB := getScore(scoresB, b)
			scoresASum += int(scoreA)
			scoresBSum += int(scoreB)
			sharedScores = append(sharedScores, [2]uint8{scoreA, scoreB})

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

func (p *Pearson) IndexesToPearson(indexA, indexB uint32, anime, manga bool) (int, float32) {
	numerator := float32(0.0)
	denominatorA := float32(0.0)
	denominatorB := float32(0.0)

	sharedScores := make([][2]uint8, 0)
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
