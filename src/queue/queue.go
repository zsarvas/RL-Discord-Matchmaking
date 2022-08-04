package queue

import (
	"strconv"
	"strings"
)

type Queue struct {
	q []string
}

func (queue *Queue) queueIsEmpty() bool {
	return len(queue.q) == 0
}

func NewQueue() *Queue {
	q := Queue{
		q: []string{},
	}

	return &q
}

func (queue *Queue) Enqueue(val string) {
	queue.q = append(queue.q, val)
}

func (queue *Queue) Dequeue() string {
	removedValue := queue.q[0]
	queue.q = queue.q[1:]

	return removedValue
}

func (queue *Queue) PlayerInQueue(player string) bool {
	for _, val := range queue.q {
		if val == player {
			return true
		}
	}

	return false
}

func (queue *Queue) LeaveQueue(player string) bool {
	newSlice := []string{}
	playerRemoved := false

	for _, val := range queue.q {
		if val == player {
			playerRemoved = true
			continue
		}
		newSlice = append(newSlice, val)
	}

	queue.q = newSlice
	return playerRemoved
}

func (queue *Queue) DisplayQueue() string {
	if queue.queueIsEmpty() {
		return ""
	}

	displayQueue := []string{}

	for idx, val := range queue.q {
		// Make a slice of wanted variables
		placement := strconv.Itoa(idx + 1)
		sanitizedName := strings.Split(val, "#")[0]
		s := []string{placement, ". ", sanitizedName}

		numberedName := strings.Join(s, "")
		displayQueue = append(displayQueue, numberedName)
	}

	currentQueue := strings.Join(displayQueue, "\n")

	return currentQueue
}
