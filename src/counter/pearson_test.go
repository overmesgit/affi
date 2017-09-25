package counter

import (
	"math"
	"testing"
	"updater"
	"math/rand"
	"fmt"
)

func TestPearson(t *testing.T) {
	// https://en.wikipedia.org/wiki/Pearson_correlation_coefficient
	pearson := NewPearson()

	user1 := updater.UserData{Id: 1, AnimeScores: []updater.UserScore{
		{Id: 1, Sc: 5},
		{Id: 2, Sc: 6},
		{Id: 3, Sc: 4},
		{Id: 4, Sc: 7},
	}, MangaScores: []updater.UserScore{
		{Id: 1, Sc: 1},
		{Id: 2, Sc: 2},
	}}

	user2 := updater.UserData{Id: 2, AnimeScores: []updater.UserScore{
		{Id: 1, Sc: 7},
		{Id: 2, Sc: 6},
		{Id: 3, Sc: 4},
		{Id: 4, Sc: 0},
	}, MangaScores: []updater.UserScore{
		{Id: 1, Sc: 2},
	}}

	index1 := pearson.UpdateUserSlices(user1)
	index2 := pearson.UpdateUserSlices(user2)

	// Check user1 and user2 pearson correlation
	scoreA := user1.AnimeScores
	avg1 := float64(scoreA[0].Sc+scoreA[1].Sc+scoreA[2].Sc) / 3.0

	scoreB := user2.AnimeScores
	avg2 := float64(scoreB[0].Sc+scoreB[1].Sc+scoreB[2].Sc) / 3.0

	numerator := (float64(scoreA[0].Sc)-avg1)*(float64(scoreB[0].Sc)-avg2) + (float64(scoreA[1].Sc)-avg1)*(float64(scoreB[1].Sc)-avg2) + (float64(scoreA[2].Sc)-avg1)*(float64(scoreB[2].Sc)-avg2)
	denominator1 := math.Pow(float64(scoreA[0].Sc)-avg1, 2) + math.Pow(float64(scoreA[1].Sc)-avg1, 2) + math.Pow(float64(scoreA[2].Sc)-avg1, 2)
	denominator2 := math.Pow(float64(scoreB[0].Sc)-avg2, 2) + math.Pow(float64(scoreB[1].Sc)-avg2, 2) + math.Pow(float64(scoreB[2].Sc)-avg2, 2)
	result := numerator / math.Sqrt(denominator1*denominator2)

	shared, pearsonResult := pearson.IndexesToPearson(index1, index2, true, false)
	if shared != 3 {
		t.Error("Wrong shared number")
	}
	if math.Abs(result-float64(pearsonResult)) > 0.001 {
		t.Error("Wrong pearson number")
	}

	shared, pearsonResult = pearson.IndexesToPearson(index1, index2, false, true)
	if shared != 1 || pearsonResult != 0 {
		t.Errorf("Wrong manga correlation %v %v", shared, pearsonResult)
	}

}

func TestScoresUpdate(t *testing.T) {
	// https://en.wikipedia.org/wiki/Pearson_correlation_coefficient
	pearson := NewPearson()

	user1 := updater.UserData{Id: 1, AnimeScores: []updater.UserScore{
		{Id: 3, Sc: 4},
		{Id: 4, Sc: 7},
	}, MangaScores: []updater.UserScore{
		{Id: 1, Sc: 1},
		{Id: 2, Sc: 2},
	}}

	user2 := updater.UserData{Id: 2, AnimeScores: []updater.UserScore{
		{Id: 1, Sc: 7},
		{Id: 2, Sc: 4},
		{Id: 3, Sc: 4},
	}, MangaScores: []updater.UserScore{
		{Id: 1, Sc: 2},
	}}
	pearson.UpdateUserSlices(user1)
	pearson.UpdateUserSlices(user2)

	prev := uint16(0)
	for _, v := range pearson.AnimeIndexes[1] {
		if v < prev {
			t.Fatal("wrong sorting")
		}
	}

}

func BenchmarkPearson(b *testing.B) {
	pearson := NewPearson()

	var firstUser updater.UserData
	for i := 0; i < 10000; i++ {
		userAnimeScores := make([]updater.UserScore, 0)
		for j := 0; j < 80; j++ {
			score := updater.UserScore{Id: uint(rand.Intn(100)), Sc: uint8(rand.Intn(10))}
			userAnimeScores = append(userAnimeScores, score)
		}

		userMangaScores := make([]updater.UserScore, 0)
		for j := 0; j < 20; j++ {
			score := updater.UserScore{Id: uint(rand.Intn(100)), Sc: uint8(rand.Intn(10))}
			userMangaScores = append(userMangaScores, score)
		}
		user := updater.UserData{Id: uint(i+1), AnimeScores: userAnimeScores, MangaScores: userMangaScores}
		pearson.UpdateUserSlices(user)
		if i == 0 {
			firstUser = user
		}
	}
	fmt.Println(len(pearson.AnimeScores))
	fmt.Println(len(pearson.AnimeIdToIndex))
	fmt.Println(len(pearson.MangaIdToIndex))
	b.ResetTimer()

	//f, err := os.Create("prof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	for i := 0; i < b.N; i++ {
		pearson.Count(firstUser, 10, true, true)
	}
}
