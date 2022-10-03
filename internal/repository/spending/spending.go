package spending

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	decimalFactor float64 = 100
)

type Spending struct {
	categories []Category
	events     []Event
	mutex      *sync.RWMutex
}

type Event struct {
	Id       int
	Category Category
	Date     time.Time
	Price    PriceFloat64
}

type PriceFloat64 int64

func (p PriceFloat64) Original() int64 {
	return int64(p)
}

func (p PriceFloat64) Float() float64 {
	return float64(p) / decimalFactor
}

func (p PriceFloat64) String() string {
	return fmt.Sprintf("%.2f", float64(p)/decimalFactor)
}

func NewSpending() *Spending {
	return &Spending{
		mutex: &sync.RWMutex{},
	}
}

func (s Spending) Events(context.Context) (e []Event) {
	s.mutex.RLock()
	e = s.events
	s.mutex.RUnlock()

	return e
}

func (s *Spending) AddEvent(ctx context.Context, categoryId int, date time.Time, price float64) ([]Event, error) {
	var category Category
	categoryFound := false
	s.mutex.RLock()
	for _, c := range s.categories {
		if c.Id == categoryId {
			category = c
			categoryFound = true
			break
		}
	}
	s.mutex.RUnlock()
	if !categoryFound {
		return nil, errors.New("category not found")
	}
	s.mutex.Lock()
	s.events = append(s.events, Event{
		Id:       genEventId(),
		Category: category,
		Date:     date,
		Price:    Float64ToPrice(price),
	})
	s.mutex.Unlock()

	return s.Events(ctx), nil
}

func (s *Spending) DeleteEvent(ctx context.Context, id int) ([]Event, error) {
	s.mutex.Lock()
	for i, event := range s.events {
		if event.Id == id {
			s.events = append(s.events[0:i], s.events[i+1:]...)
			break
		}
	}
	s.mutex.Unlock()

	return s.Events(ctx), nil
}

func (s Spending) Report(ctx context.Context, f1, f2 time.Time) (m map[int]float64) {
	_ = ctx

	s.mutex.RLock()
	stat := make(map[int]PriceFloat64)
	for _, event := range s.events {
		if (event.Date.After(f1) || event.Date.Equal(f1)) && (event.Date.Before(f2) || event.Date.Equal(f2)) {
			stat[event.Category.Id] += event.Price
		}
	}
	s.mutex.RUnlock()
	m = make(map[int]float64)
	for catId, price := range stat {
		m[catId] = price.Float()
	}

	return m
}

var genEventId = func() func() int {
	c := -1
	return func() int {
		c++
		return c
	}
}()

func Float64ToPrice(f float64) PriceFloat64 {
	return PriceFloat64(f * decimalFactor)
}
