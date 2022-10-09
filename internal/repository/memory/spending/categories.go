package spending

import (
	"context"
	"strings"
)

type Category struct {
	Id    int
	Title string
}

func (s Spending) Categories(context.Context) (c []Category) {
	s.mutex.RLock()
	c = s.categories
	s.mutex.RUnlock()

	return
}

func (s *Spending) AddCategory(ctx context.Context, title string) ([]Category, error) {
	s.mutex.Lock()
	categoryFound := false
	for _, category := range s.categories {
		if strings.EqualFold(category.Title, title) {
			categoryFound = true
			break
		}
	}
	if categoryFound {
		s.mutex.Unlock()
		return s.Categories(ctx), nil
	}
	s.categories = append(s.categories, Category{
		Id:    genCategoryId(),
		Title: title,
	})
	s.mutex.Unlock()

	return s.Categories(ctx), nil
}

func (s *Spending) DeleteCategory(ctx context.Context, id int) ([]Category, error) {
	s.mutex.Lock()
	for i, category := range s.categories {
		if category.Id == id {
			s.categories = append(s.categories[0:i], s.categories[i+1:]...)
			break
		}
	}
	s.mutex.Unlock()

	return s.Categories(ctx), nil
}

var genCategoryId = func() func() int {
	c := -1
	return func() int {
		c++
		return c
	}
}()
