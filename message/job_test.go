package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLinkedTaskByMinTime(t *testing.T) {
	testCase := []struct {
		time []time.Duration
	}{
		{
			time: []time.Duration{
				time.Second,
				time.Second * 2,
				time.Second * 3,
				time.Second * 4,
			},
		},
		{
			time: []time.Duration{
				time.Second,
				time.Second * 2,
				time.Second * 4,
				time.Second * 3,
			},
		},
		{
			time: []time.Duration{
				time.Second,
				time.Second * 2,
				- time.Second,
				time.Second * 3,
			},
		},
	}

	for _, v := range testCase {
		cron := newCron(time.Second)
		now := time.Now()

		for _, duration := range v.time {
			cron.add(messageSet{}, now.Add(duration))
		}

		total := 0
		prevTask := &messageTask{}

		for task := cron.head; task != nil; task = task.next {
			if prevTask.unix > task.unix {
				t.Fatal("wrong task order")
			}
			prevTask = task
			total++
		}

		assert.Equal(t, cron.len, total)
	}
}

func TestAddTheSameNode(t *testing.T) {
	cron := newCron(time.Second)
	now := time.Now()
	cron.add(messageSet{message: Message{Id: 1}}, now)
	cron.add(messageSet{message: Message{Id: 2}}, now.Add(time.Second))
	cron.add(messageSet{message: Message{Id: 3}}, now)
	cron.add(messageSet{message: Message{Id: 4}}, now.Add(time.Second))

	var id []int64
	for node := cron.head; node != nil; node = node.next {
		for _, v := range node.message {
			id = append(id, v.message.Id)
		}
	}

	assert.Equal(t, []int64{1, 3, 2, 4}, id)
}

func TestPop(t *testing.T) {
	cron := newCron(time.Second)
	now := time.Now()
	cron.add(messageSet{}, now)
	cron.add(messageSet{}, now.Add(time.Second))
	cron.pop()

	assert.Equal(t, now.Add(time.Second), cron.head.time)
	assert.Equal(t, 1, cron.len)

	cron.pop()

	assert.Nil(t, cron.head)
	assert.Equal(t, 0, cron.len)
}

func TestStart(t *testing.T) {
	cron := newCron(time.Second)

	cron.add(messageSet{message: Message{Id: 1}}, time.Now().Add(time.Second))
	cron.add(messageSet{message: Message{Id: 2}}, time.Now().Add(time.Second*2))
	cron.add(messageSet{message: Message{Id: 3}}, time.Now().Add(time.Second*3))

	cron.start()
	var msg []messageSet
	tc := time.After(time.Second * 4)

	go func() {
		for {
			select {
			case m := <-cron.Message():
				msg = append(msg, m...)
			case <-tc:
				break
			}
		}
	}()

	<-tc

	assert.Equal(t, 1, cron.len)
	assert.Equal(t, []messageSet{
		messageSet{message: Message{Id: 1}},
		messageSet{message: Message{Id: 2}},
	}, msg)
}
