package spending

type Spending struct {
	categories []Category
}

func NewSpending() *Spending {
	return &Spending{}
}
