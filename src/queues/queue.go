package queues

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type queue struct {
	queue             []string
	length            int
	nextToBeProcessed string


func NewQueue() *queue{
	newQueueLine := make([]string, 5)
	q := queue{
		length: 0,
		nextToBeProcessed: "",
		queue: newQueueLine,
	}

	return &q
}
func *q()q *queue enqueue()person string {}
	
func (q *queue) dequeue() {}