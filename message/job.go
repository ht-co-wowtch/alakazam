package message

import (
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 一個單向 Linked List結構
type cron struct {
	len   int
	head  *messageTask
	msg   chan []messageSet
	stop  chan struct{}
	rate  time.Duration
	lock  sync.Mutex
	isRun bool
}

func newCron(rate time.Duration) *cron {
	return &cron{
		stop: make(chan struct{}),
		msg:  make(chan []messageSet, 100),
		rate: rate,
	}
}

func (c *cron) close() {
	c.stop <- struct{}{}
	c.isRun = false
}

func (c *cron) run() {
	c.isRun = true
	t := time.NewTicker(c.rate)
	for {
		select {
		case now := <-t.C:
			if c.head == nil {
				continue
			}
			if c.head.unix < now.Unix() {
				select {
				case c.msg <- c.head.message:
				default:
					log.Warnf("message miss for cron", zap.Int64s("id", c.head.Ids()))
				}
				c.pop()
			}
		case <-c.stop:
			return
		}
	}
}

func (c *cron) Message() <-chan []messageSet {
	return c.msg
}

// 新增一個task至一個Linked List，排序方式以最小時間
// 時間越小則Linked List越前面，反之則最後面
func (c *cron) add(message messageSet, time time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()

	task := newMessageTask([]messageSet{message}, time)
	// Linked 沒有資料時
	if c.head == nil {
		c.head = task
		c.len++
		// 當Linked第一筆資料時間比task時間還要大則該task要優先執行
	} else if c.head.unix > task.unix {
		task.next = c.head
		c.head.prev = task
		c.head = task
		c.len++
	} else {
		for node := c.head; node != nil; node = node.next {
			// 如果node時間都一樣則放在一起
			if node.unix == task.unix {
				node.add(message)
				c.len++
				return
			}

			// 該task資料時間為Linked中最小
			if node.next == nil {
				task.prev = node
				node.next = task
				c.len++
				return
			}
			// 該task資料時間介於兩個node之間
			if node.unix < task.unix && task.unix < node.next.unix {
				task.prev = node
				task.next = node.next
				node.next = task
				node.next.prev = task
				c.len++
				return
			}
		}
	}
}

func (c *cron) pop() {
	c.lock.Lock()
	task := c.head.next
	if task != nil {
		task.prev = nil
		c.head = task
	} else {
		c.head = nil
	}
	c.len--
	c.lock.Unlock()
}

type messageTask struct {
	message []messageSet
	time    time.Time
	unix    int64
	prev    *messageTask
	next    *messageTask
}

func newMessageTask(message []messageSet, time time.Time) *messageTask {
	return &messageTask{
		message: message,
		time:    time,
		unix:    time.Unix(),
	}
}

func (m *messageTask) add(message messageSet) {
	m.message = append(m.message, message)
}

func (m *messageTask) Ids() []int64 {
	id := make([]int64, 0, len(m.message))
	for _, v := range m.message {
		id = append(id, v.message.Id)
	}
	return id
}

const (
	message_category             = 1
	redenvelope_message_category = 2
)

type messageSet struct {
	room        []string
	message     Message
	redEnvelope RedEnvelope
	category    int
}
