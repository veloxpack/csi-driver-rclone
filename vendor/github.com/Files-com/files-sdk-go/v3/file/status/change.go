package status

import "time"

type Change struct {
	Status
	time.Time
	Err error
}

type Changes []Change

func (c Changes) Count(s Status) (count int) {
	for _, change := range c {
		if change.Is(s) {
			count += 1
		}
	}

	return count
}
