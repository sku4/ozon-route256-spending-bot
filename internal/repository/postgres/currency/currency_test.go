package currency

import (
	"github.com/jmoiron/sqlx"
	"github.com/sku4/ozon-route256-spending-bot/model"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
)

func TestNewCurrencies(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	tests := []struct {
		name    string
		mock    func()
		want    []model.Currency
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "abbreviation"}).
					AddRow(1, "USD").
					AddRow(2, "CNY")
				mock.ExpectQuery("SELECT (.+) FROM currency").
					WillReturnRows(rows)
			},
			want: []model.Currency{
				{
					Id:   1,
					Abbr: "USD",
				},
				{
					Id:   2,
					Abbr: "CNY",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := NewCurrencies(db)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.currencies)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
