package queue

import "strings"

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

func (queue *Queue) DisplayQueue() string {
	if queue.queueIsEmpty() {
		return ""
	}

	displayQueue := []string{}

	for _, val := range queue.q {
		santizedName := strings.Split(val, "#")[0]
		displayQueue = append(displayQueue, santizedName)
	}

	currentQueue := strings.Join(displayQueue, ", ")

	return currentQueue
}
