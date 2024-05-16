package currency

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"
)

type Currency struct {
	Name           string        `json:"name"`
	Value          int64         `json:"-"`
	ValueString    string        `json:"value"`
	Decimal        uint8         `json:"decimal"`
	ChangeMin      int64         `json:"-"`
	ChangeMax      int64         `json:"-"`
	Interval       time.Duration `json:"-"`
	IntervalString int           `json:"interval_in_ms"`
}

func (c Currency) GetNextJSON() ([]byte, error) {
	newChange := randRange(c.ChangeMin, c.ChangeMax)
	// Currency shall not go negative
	if newChange < 0 && c.Value+newChange <= 0 {
		c.Value = c.Value - newChange
	} else {
		c.Value = c.Value + newChange
	}
	c.ValueString = strconv.FormatInt(c.Value, 10)
	return json.Marshal(c)
}

// min is inclusive; max is exclusive
func randRange(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

var Currencies = []Currency{
	{
		Name:           "Nani",
		Value:          123456789012345678,
		Decimal:        8,
		ChangeMin:      -12345678,
		ChangeMax:      12345678,
		Interval:       10 * time.Second,
		IntervalString: 1000,
	},
	{
		Name:           "Programming",
		Value:          999888777666,
		Decimal:        6,
		ChangeMin:      -555444333,
		ChangeMax:      555444333,
		Interval:       5 * time.Second,
		IntervalString: 5000,
	},
	{
		Name:           "Is",
		Value:          987654321098765432,
		Decimal:        12,
		ChangeMin:      -12345,
		ChangeMax:      12345,
		Interval:       800 * time.Millisecond,
		IntervalString: 800,
	},
	{
		Name:           "Fun",
		Value:          11111111111,
		Decimal:        11,
		ChangeMin:      -1111,
		ChangeMax:      1111,
		Interval:       1111 * time.Millisecond,
		IntervalString: 1111,
	},
}
