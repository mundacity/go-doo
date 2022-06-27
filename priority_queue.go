package godoo

import (
	"container/heap"
)

// examples here:
// - https://pkg.go.dev/container/heap
// - https://github.com/andreimcristof/go-starter/blob/0599b5af8307338c7ea15728bf50fa16de8a07e6/datastructures-queues-and-heaps/patients_queue.go

// PriorityQueue implements heap interface methods
type PriorityQueue struct {
	DateMode bool
	Items    map[int]*TodoItem
}

// indexMap tracks which indice relates to which map key
var indexMap map[int]int

// NewPriorityQueue constructor for PriorityQueue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{DateMode: false, Items: make(map[int]*TodoItem)}
	indexMap = make(map[int]int)
	heap.Init(pq)
	return pq
}

// Len required by heap.Interface
func (pq PriorityQueue) Len() int {
	return len(pq.Items)
}

// Less required by heap.Interface
func (pq PriorityQueue) Less(a, b int) bool {
	if !pq.DateMode {
		aKey := pq.getMapKeyFromIndex(a)
		bKey := pq.getMapKeyFromIndex(b)
		b := pq.Items[aKey].Priority > pq.Items[bKey].Priority
		return b
	} else {
		return pq.Items[a].Deadline.After(pq.Items[b].Deadline)
	}
}

// Pop required by heap.Interface
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.Items
	n := len(old)
	nKey := pq.getMapKeyFromIndex(n - 1)
	itm := old[nKey]
	k := itm.Index
	old[nKey] = nil // apparently helps to avoid memory leaks --> in both slice-based implementations (not sure if necessary for map-based implementation)
	itm.Index = -1  // apparently for safety
	delete(pq.Items, itm.Id)
	delete(indexMap, k)
	return itm
}

// Push required by heap.Interface
func (pq *PriorityQueue) Push(todoItm interface{}) {
	itm := todoItm.(TodoItem)
	itm.Index = len(pq.Items) // required by heap.interface
	pq.Items[itm.Id] = &itm
	addToIndexMap(itm.Index, itm.Id)
}

// Swap required by heap.Interface
func (pq *PriorityQueue) Swap(a, b int) {
	aKey := pq.getMapKeyFromIndex(a)
	bKey := pq.getMapKeyFromIndex(b)

	indexMap[a], indexMap[b] = indexMap[b], indexMap[a]
	pq.Items[aKey].Index = b
	pq.Items[bKey].Index = a
}

func addToIndexMap(idx, key int) {
	indexMap[idx] = key
}

func (pq PriorityQueue) getMapKeyFromIndex(idx int) int {
	return indexMap[idx]
}
