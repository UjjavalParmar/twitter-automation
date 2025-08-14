package scheduler

import (
	"fmt"
	"math/rand"
	"time"
)

type Slot struct {
	Time time.Time
	Key  string // yyyymmdd-HHMM
}

func DailyRandomSlots(loc *time.Location, n int, startHHMM, endHHMM string) ([]Slot, error) {
	start, err := parseHHMM(startHHMM)
	if err != nil {
		return nil, err
	}
	end, err := parseHHMM(endHHMM)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, fmt.Errorf("window end must be after start")
	}

	var slots []Slot
	day := time.Now().In(loc)
	for i := 0; i < n; i++ {
		// random minute offset within window
		delta := time.Duration(rand.Int63n(int64(end.Sub(start))))
		t := time.Date(day.Year(), day.Month(), day.Day(), start.Hour(), start.Minute(), 0, 0, loc).Add(delta)
		key := t.Format("20060102-1504")
		slots = append(slots, Slot{Time: t, Key: key})
	}
	// sort
	for i := range slots {
		for j := i + 1; j < len(slots); j++ {
			if slots[j].Time.Before(slots[i].Time) {
				slots[i], slots[j] = slots[j], slots[i]
			}
		}
	}
	return slots, nil
}

func parseHHMM(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Time{}, err
	}
	// zero date; only duration matters
	return t, nil
}
