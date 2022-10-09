package spending

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestSpending_AddDeleteEvents(t *testing.T) {
	ctx := context.Background()
	t.Run("", func(t *testing.T) {
		s := NewSpending()
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		gotCategories, _ := s.AddCategory(ctx, "category 1")
		category := gotCategories[0]
		_, _ = s.AddEvent(ctx, category.Id, now, 15.6)
		got, _ := s.AddEvent(ctx, category.Id, next, 50)
		wantEvents := []Event{
			{
				Id:       0,
				Category: category,
				Date:     now,
				Price:    Float64ToPrice(15.6),
			},
			{
				Id:       1,
				Category: category,
				Date:     next,
				Price:    Float64ToPrice(50),
			},
		}
		if !reflect.DeepEqual(got, wantEvents) {
			t.Errorf("AddEvent() got = %v, want %v", got, wantEvents)
		}
		got, _ = s.DeleteEvent(ctx, 0)
		wantEvents = []Event{
			{
				Id:       1,
				Category: category,
				Date:     next,
				Price:    Float64ToPrice(50),
			},
		}
		if !reflect.DeepEqual(got, wantEvents) {
			t.Errorf("DeleteEvent() got = %v, want %v", got, wantEvents)
		}
	})
}
