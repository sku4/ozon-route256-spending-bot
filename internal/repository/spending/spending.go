package spending

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type Spending struct {
	categories []Category
	events     []Event
}

type Event struct {
	Id       int
	Category Category
	Date     time.Time
	Price    float64
}

func NewSpending() *Spending {
	return &Spending{}
}

func (s Spending) Events(context.Context) []Event {
	return s.events
}

func (s *Spending) AddEvent(ctx context.Context, categoryId int, date time.Time, price float64) ([]Event, error) {
	var category Category
	categoryFound := false
	for _, c := range s.categories {
		if c.Id == categoryId {
			category = c
			categoryFound = true
			break
		}
	}
	if !categoryFound {
		return nil, errors.New("category not found")
	}
	s.events = append(s.events, Event{
		Id:       genEventId(),
		Category: category,
		Date:     date,
		Price:    price,
	})

	return s.Events(ctx), nil
}

func (s *Spending) DeleteEvent(ctx context.Context, id int) ([]Event, error) {
	for i, event := range s.events {
		if event.Id == id {
			s.events = append(s.events[0:i], s.events[i+1:]...)
			break
		}
	}

	return s.Events(ctx), nil
}

var genEventId = func() func() int {
	c := -1
	return func() int {
		c++
		return c
	}
}()
