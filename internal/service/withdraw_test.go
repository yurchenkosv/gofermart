package service

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/yurchenkosv/gofermart/internal/dao"
	mock_dao "github.com/yurchenkosv/gofermart/internal/mocks"
	"github.com/yurchenkosv/gofermart/internal/model"
	"testing"
	"time"
)

func TestGetWithdrawalsForCurrentUser(t *testing.T) {
	type mockBehavior func(s *mock_dao.MockRepository, withdraw model.Withdraw, id int)
	type args struct {
		withdraw model.Withdraw
	}
	tests := []struct {
		name     string
		id       int
		args     args
		behavior mockBehavior
		want     []*model.Withdraw
		wantErr  bool
	}{
		{
			behavior: func(s *mock_dao.MockRepository, withdraw model.Withdraw, id int) {
				s.EXPECT().GetWithdrawals(withdraw).Return([]*model.Withdraw{
					{
						Order:       "12345",
						Sum:         500,
						ProcessedAt: time.Unix(172386238, 13),
						User: model.User{
							ID: &id,
						},
					},
				}, nil)
			},
			id: 1,
			want: []*model.Withdraw{
				{
					Order:       "12345",
					Sum:         500,
					ProcessedAt: time.Unix(172386238, 13),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authRepo := mock_dao.NewMockRepository(ctrl)
			withdrawService := NewWithdrawService(authRepo)
			tt.behavior(authRepo, tt.args.withdraw, tt.id)
			got, err := withdrawService.GetWithdrawalsForCurrentUser(tt.args.withdraw)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tt.want[0].Order, got[0].Order)
		})
	}
}

func TestProcessWithdraw(t *testing.T) {
	type args struct {
		withdraw   model.Withdraw
		repository dao.Repository
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, true)
		})
	}
}
