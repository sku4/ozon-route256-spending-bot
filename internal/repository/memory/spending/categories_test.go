package spending

import (
	"context"
	"reflect"
	"testing"
)

func TestSpending_AddDeleteCategories(t *testing.T) {
	ctx := context.Background()
	t.Run("", func(t *testing.T) {
		s := NewSpending()
		got, _ := s.AddCategory(ctx, "category 1")
		wantCategories := []Category{
			{
				Id:    0,
				Title: "category 1",
			},
		}
		if !reflect.DeepEqual(got, wantCategories) {
			t.Errorf("AddCategory() got = %v, want %v", got, wantCategories)
		}
		_, _ = s.AddCategory(ctx, "category 2")
		got, _ = s.DeleteCategory(ctx, 0)
		wantCategories = []Category{
			{
				Id:    1,
				Title: "category 2",
			},
		}
		if !reflect.DeepEqual(got, wantCategories) {
			t.Errorf("DeleteCategory() got = %v, want %v", got, wantCategories)
		}
	})
}
