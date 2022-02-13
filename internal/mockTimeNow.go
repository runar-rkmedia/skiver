package internal

import (
	"sync"
	"time"

	"bou.ke/monkey"
)

type IncreasingTimeMock struct {
	Ticks       time.Duration
	TickAmmount time.Duration
	sync.RWMutex
	StartTime time.Time
}

func NewMockTimeNow() *IncreasingTimeMock {
	mt := IncreasingTimeMock{Ticks: 0, TickAmmount: time.Second}
	t, err := time.Parse("2006-01-02", "1987-01-01")
	if err != nil {
		panic(err)
	}
	mt.StartTime = t
	monkey.Patch(time.Now, mt.Now)
	return &mt
}

func (mt *IncreasingTimeMock) Tick() time.Time {
	mt.Lock()
	defer mt.Unlock()
	now := mt.StartTime.Add(mt.TickAmmount * mt.Ticks)
	mt.Ticks += 1
	return now

}
func (mt *IncreasingTimeMock) Now() time.Time {
	mt.RLock()
	defer mt.RUnlock()
	now := mt.StartTime.Add(mt.TickAmmount * mt.Ticks)
	return now
}
