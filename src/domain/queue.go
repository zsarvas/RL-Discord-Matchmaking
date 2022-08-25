package domain

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Queue struct {
	queue     []Player
	popLength int
}

func NewQueue(popsAt int) *Queue {
	q := Queue{
		queue:     []Player{},
		popLength: popsAt,
	}

	return &q
}

func (queue *Queue) queueIsEmpty() bool {
	return len(queue.queue) == 0
}

func (queue *Queue) Enqueue(player Player) bool {
	queue.queue = append(queue.queue, player)

	return len(queue.queue) == queue.popLength
}

func (queue *Queue) Dequeue() Player {
	removedValue := queue.queue[0]
	queue.queue = queue.queue[1:]

	return removedValue
}

func (queue *Queue) PlayerInQueue(player Player) bool {
	for _, queuePlayer := range queue.queue {
		if queuePlayer.Id == player.Id {
			return true
		}
	}

	return false
}

func (queue *Queue) LeaveQueue(player Player) bool {
	newSlice := []Player{}
	playerRemoved := false

	for _, queuePlayer := range queue.queue {
		if queuePlayer.Id == player.Id {
			playerRemoved = true
			continue
		}
		newSlice = append(newSlice, queuePlayer)
	}

	queue.queue = newSlice
	return playerRemoved
}

func (queue *Queue) DisplayQueue() string {
	if queue.queueIsEmpty() {
		return ""
	}

	displayQueue := []string{}

	for idx, player := range queue.queue {
		// Make a slice of wanted variables
		placement := strconv.Itoa(idx + 1)
		s := []string{placement, ". ", player.MentionName}

		numberedName := strings.Join(s, "")
		displayQueue = append(displayQueue, numberedName)
	}

	currentQueue := strings.Join(displayQueue, "\n")

	return currentQueue
}

func (queue *Queue) GetPopLength() int {
	return queue.popLength
}

func (queue *Queue) GetQueueLength() int {
	return len(queue.queue)
}

func (queue *Queue) RandomizeQueue() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(queue.queue), func(i, j int) { queue.queue[i], queue.queue[j] = queue.queue[j], queue.queue[i] })
}
