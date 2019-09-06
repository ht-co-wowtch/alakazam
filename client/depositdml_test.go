package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestGetDepositAndDmlByTime(t *testing.T) {
	c := newMockDepositAndDmlClient(t, 1)
	_, err := c.GetDepositAndDml(1, "", "")

	assert.Nil(t, err)

	c = newMockDepositAndDmlClient(t, 2)
	_, err = c.GetDepositAndDml(2, "", "")

	assert.Nil(t, err)
}

func newMockDepositAndDmlClient(t *testing.T, day int) *Client {
	return newMockClient(func(req *http.Request) (resp *http.Response, err error) {
		query := req.URL.Query()
		start, end, err := getTimeRange(query.Get("start_at"), query.Get("end_at"))
		if err != nil {
			t.Fatal(err)
		}

		today, err := getMidnight(time.Now().AddDate(0, 0, -(day - 1)))
		if err != nil {
			t.Fatal(err)
		}

		day := int(end.Sub(start).Hours())
		diff := int(time.Now().Sub(today).Hours())
		if day != diff {
			t.Fatalf("not one day，got %d ， expected %d", day, diff)
		}

		body, err := json.Marshal(Money{})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		}, nil
	})
}

func getTimeRange(startAt, endAt string) (start time.Time, end time.Time, err error) {
	start, err = time.Parse(time.RFC3339, startAt)
	if err != nil {
		return start, end, fmt.Errorf("start_at time parse %s", err.Error())
	}
	end, err = time.Parse(time.RFC3339, endAt)
	if err != nil {
		return start, end, fmt.Errorf("end_at time parse %s", err.Error())
	}
	return start, end, nil
}

func getMidnight(day time.Time) (time.Time, error) {
	return time.Parse(time.RFC3339, day.Format("2006-01-02")+"T00:00:00+08:00")
}
