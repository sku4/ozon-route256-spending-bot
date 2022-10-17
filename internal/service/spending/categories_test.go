//go:build integration
// +build integration

package spending

import (
	"context"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
)

func TestService_CategoryAdd(t *testing.T) {
	ctx := context.Background()
	st, s, mock, err := NewServiceTest(ctx)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	tests := []struct {
		name     string
		mockFunc func()
		title    string
		command  string
		wantErr  bool
	}{
		{
			name: "Ok",
			mockFunc: func() {
				rowsCategory := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("INSERT INTO category").
					WithArgs("Food").WillReturnRows(rowsCategory)
			},
			title:   "Food",
			command: "/categoryadd",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			tgBotUpdateCommand := st.TgBotMessageCommand(tt.command, tt.title)
			err = s.CategoryAdd(ctx, tgBotUpdateCommand)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
