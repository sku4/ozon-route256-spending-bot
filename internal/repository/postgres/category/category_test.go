package category

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/sku4/ozon-route256-spending-bot/model"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
)

func TestCategory_Categories(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	r := NewCategory(db)

	ctx := context.Background()
	tests := []struct {
		name    string
		mock    func()
		want    []model.Category
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title"}).
					AddRow(1, "Food").
					AddRow(2, "Auto")
				mock.ExpectQuery("SELECT (.+) FROM category").
					WillReturnRows(rows)
			},
			want: []model.Category{
				{
					Id:    1,
					Title: "Food",
				},
				{
					Id:    2,
					Title: "Auto",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.Categories(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCategory_AddCategory(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	r := NewCategory(db)

	ctx := context.Background()
	tests := []struct {
		name           string
		mock           func()
		title          string
		wantCategoryId int
		wantErr        bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("INSERT INTO category").
					WithArgs("Food").WillReturnRows(rows)
			},
			title:          "Food",
			wantCategoryId: 1,
			wantErr:        false,
		},
		{
			name: "Empty Fields",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"})
				mock.ExpectQuery("INSERT INTO category").
					WithArgs("Auto").WillReturnRows(rows)
			},
			title:   "Auto",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.AddCategory(ctx, tt.title)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCategoryId, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCategory_DeleteCategory(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	r := NewCategory(db)
	ctx := context.Background()
	tests := []struct {
		name    string
		mock    func()
		id      int
		wantErr bool
	}{

		{
			name: "Ok",
			mock: func() {
				mock.ExpectExec("DELETE FROM category WHERE (.+)").
					WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			id:      1,
			wantErr: false,
		},
		{
			name: "Not Found",
			mock: func() {
				mock.ExpectExec("DELETE FROM category WHERE (.+)").
					WithArgs(2).WillReturnError(sql.ErrNoRows)
			},
			id:      2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err = r.DeleteCategory(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCategory_CategoryGetById(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	r := NewCategory(db)

	ctx := context.Background()
	tests := []struct {
		name    string
		id      int
		mock    func()
		want    *model.Category
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title"}).
					AddRow(2, "Auto")
				mock.ExpectQuery("SELECT (.+) FROM category WHERE (.+)").
					WillReturnRows(rows)
			},
			id: 2,
			want: &model.Category{
				Id:    2,
				Title: "Auto",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.CategoryGetById(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCategory_CategoryGetByTitle(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	r := NewCategory(db)

	ctx := context.Background()
	tests := []struct {
		name    string
		title   string
		mock    func()
		want    *model.Category
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title"}).
					AddRow(2, "Auto")
				mock.ExpectQuery("SELECT (.+) FROM category WHERE (.+)").
					WillReturnRows(rows)
			},
			title: "Auto",
			want: &model.Category{
				Id:    2,
				Title: "Auto",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.CategoryGetByTitle(ctx, tt.title)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
