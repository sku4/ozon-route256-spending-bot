package spending

import (
	"context"
	"strings"
)

type Category struct {
	Id    int
	Title string
}

func (s Spending) Categories(context.Context) []Category {
	return s.categories
}

func (s *Spending) AddCategory(ctx context.Context, title string) ([]Category, error) {
	for _, category := range s.categories {
		if strings.EqualFold(category.Title, title) {
			return s.Categories(ctx), nil
		}
	}
	s.categories = append(s.categories, Category{
		Id:    genCategoryId(),
		Title: title,
	})

	return s.Categories(ctx), nil
}

func (s *Spending) DeleteCategory(ctx context.Context, id int) ([]Category, error) {
	for i, category := range s.categories {
		if category.Id == id {
			s.categories = append(s.categories[0:i], s.categories[i+1:]...)
			break
		}
	}

	return s.Categories(ctx), nil
}

var genCategoryId = func() func() int {
	c := -1
	return func() int {
		c++
		return c
	}
}()
