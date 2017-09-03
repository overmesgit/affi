package counter

type PearsonResultHeap []PearsonResult

func (h PearsonResultHeap) Len() int           { return len(h) }
func (h PearsonResultHeap) Less(i, j int) bool { return h[i].Pearson < h[j].Pearson }
func (h PearsonResultHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PearsonResultHeap) Push(x interface{}) {
	*h = append(*h, x.(PearsonResult))
}

func (h *PearsonResultHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
