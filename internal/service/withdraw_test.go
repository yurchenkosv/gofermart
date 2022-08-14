package service

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/yurchenkosv/gofermart/internal/errors"
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

func TestWithdrawService_ProcessWithdraw(t *testing.T) {
	type fields struct {
		repo *mock_dao.MockRepository
	}
	type args struct {
		withdraw model.Withdraw
	}
	tests := []struct {
		name        string
		prepare     func(f *fields, withdraw model.Withdraw)
		args        args
		wantErr     assert.ErrorAssertionFunc
		wantErrType error
	}{
		{
			name: "should success process withdraw",
			prepare: func(f *fields, withdraw model.Withdraw) {
				b := model.Balance{User: withdraw.User}
				f.repo.EXPECT().GetBalance(b).Return(
					&model.Balance{
						User:         b.User,
						Balance:      100,
						SpentAllTime: 100,
					},
					nil)
				f.repo.EXPECT().Save(&model.Balance{
					User:         withdraw.User,
					Balance:      50,
					SpentAllTime: 150,
				}).Return(nil)
				f.repo.EXPECT().Save(&withdraw).Return(nil)
			},
			wantErr:     assert.NoError,
			wantErrType: nil,
			args: args{withdraw: model.Withdraw{
				Order:       "2377225624",
				Sum:         50,
				ProcessedAt: time.Unix(123123132, 0),
				User: model.User{
					ID: GetIntPointer(1),
				},
			}},
		},
		{
			name:        "should return OrderFormatError",
			prepare:     func(f *fields, withdraw model.Withdraw) {},
			wantErr:     assert.Error,
			wantErrType: &errors.OrderFormatError{},
			args: args{withdraw: model.Withdraw{
				Order:       "1234567890",
				Sum:         50,
				ProcessedAt: time.Unix(123123132, 0),
				User: model.User{
					ID: GetIntPointer(1),
				},
			}},
		},
		{
			name: "should return LowBalanceError",
			prepare: func(f *fields, withdraw model.Withdraw) {
				b := model.Balance{User: withdraw.User}
				f.repo.EXPECT().GetBalance(b).Return(
					&model.Balance{
						User:         b.User,
						Balance:      100,
						SpentAllTime: 100,
					},
					nil)
			},
			wantErr:     assert.Error,
			wantErrType: &errors.LowBalanceError{},
			args: args{withdraw: model.Withdraw{
				Order:       "2377225624",
				Sum:         150,
				ProcessedAt: time.Unix(123123132, 0),
				User: model.User{
					ID: GetIntPointer(1),
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			f := fields{repo: mock_dao.NewMockRepository(ctrl)}
			tt.prepare(&f, tt.args.withdraw)
			s := WithdrawService{
				repo: f.repo,
			}
			err := s.ProcessWithdraw(tt.args.withdraw)
			tt.wantErr(t, err, fmt.Sprintf("ProcessWithdraw(%v)", tt.args.withdraw))
			assert.IsType(t, tt.wantErrType, err)
		})
	}
}
