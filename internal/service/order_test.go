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

func TestOrderService_CreateOrder(t *testing.T) {
	type fields struct {
		repo *mock_dao.MockRepository
	}
	type args struct {
		order *model.Order
	}
	tests := []struct {
		name          string
		prepare       func(f *fields, order *model.Order)
		args          args
		wantErr       assert.ErrorAssertionFunc
		wantErrorType error
	}{
		{
			name: "should success create order",
			prepare: func(f *fields, order *model.Order) {
				id := 1
				order.User.ID = &id
				gomock.InOrder(
					f.repo.EXPECT().GetOrderByNumber(order.Number).Return(order, nil),
					f.repo.EXPECT().Save(order).Return(nil),
				)
			},
			args: args{order: &model.Order{
				User:       &model.User{ID: new(int)},
				Number:     "2377225624",
				Status:     "NEW",
				UploadTime: time.Unix(176237653, 0),
			}},
			wantErr:       assert.NoError,
			wantErrorType: nil,
		},
		{
			name: "should return OrderAlreadyAcceptedCurrentUserError",
			prepare: func(f *fields, order *model.Order) {
				id := 1
				order.ID = &id
				order.User.ID = &id
				gomock.InOrder(
					f.repo.EXPECT().GetOrderByNumber(order.Number).Return(order, nil),
				)
			},
			args: args{order: &model.Order{
				User:       &model.User{ID: new(int)},
				Number:     "2377225624",
				Status:     "NEW",
				UploadTime: time.Unix(176237653, 0),
			}},
			wantErr:       assert.Error,
			wantErrorType: &errors.OrderAlreadyAcceptedCurrentUserError{},
		},
		{
			name: "should return OrderAlreadyAcceptedDifferentUserError",
			prepare: func(f *fields, order *model.Order) {
				orderId := 1
				userId := 2
				order.ID = &orderId
				orderInDB := *order
				orderInDB.User = &model.User{ID: &userId}
				gomock.InOrder(
					f.repo.EXPECT().GetOrderByNumber(order.Number).Return(&orderInDB, nil),
				)
			},
			args: args{order: &model.Order{
				User:       &model.User{ID: new(int)},
				Number:     "2377225624",
				Status:     "NEW",
				UploadTime: time.Unix(176237653, 0),
			}},
			wantErr:       assert.Error,
			wantErrorType: &errors.OrderAlreadyAcceptedDifferentUserError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			f := fields{repo: mock_dao.NewMockRepository(ctrl)}
			tt.prepare(&f, tt.args.order)
			s := OrderService{
				repo: f.repo,
			}
			orderService := s.CreateOrder(tt.args.order)
			tt.wantErr(t, orderService, fmt.Sprintf("CreateOrder(%v)", tt.args.order))
			assert.IsType(t, tt.wantErrorType, orderService)
		})
	}
}
func GetIntPointer(value int) *int {
	return &value
}
func TestOrderService_GetUploadedOrdersForUser(t *testing.T) {
	type fields struct {
		repo *mock_dao.MockRepository
	}
	type args struct {
		order *model.Order
	}
	tests := []struct {
		name        string
		prepare     func(f *fields, order *model.Order)
		args        args
		want        []model.Order
		wantErr     assert.ErrorAssertionFunc
		wantErrType error
	}{
		{
			name: "should success return orders for current user",
			prepare: func(f *fields, order *model.Order) {
				id := 1
				order.ID = &id
				order.User.ID = &id
				gomock.InOrder(
					f.repo.EXPECT().GetOrdersForUser(*order).Return([]model.Order{
						{
							ID:         order.ID,
							User:       order.User,
							Number:     "2377225624",
							Accrual:    nil,
							Status:     "NEW",
							UploadTime: time.Unix(123123134, 0),
						},
					}, nil),
				)
			},
			args: args{order: &model.Order{User: &model.User{ID: new(int)}}},
			want: []model.Order{
				{
					ID:         GetIntPointer(1),
					User:       &model.User{ID: GetIntPointer(1)},
					Number:     "2377225624",
					Accrual:    nil,
					Status:     "NEW",
					UploadTime: time.Unix(123123134, 0),
				},
			},
			wantErr:     assert.NoError,
			wantErrType: nil,
		},
		{
			name: "should return NoOrdersError",
			prepare: func(f *fields, order *model.Order) {
				id := 1
				order.User.ID = &id
				gomock.InOrder(
					f.repo.EXPECT().GetOrdersForUser(*order).Return([]model.Order{}, nil),
				)
			},
			args:        args{order: &model.Order{User: &model.User{ID: new(int)}}},
			want:        nil,
			wantErr:     assert.Error,
			wantErrType: &errors.NoOrdersError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			f := fields{repo: mock_dao.NewMockRepository(ctrl)}
			tt.prepare(&f, tt.args.order)
			s := OrderService{
				repo: f.repo,
			}
			got, err := s.GetUploadedOrdersForUser(tt.args.order)
			if !tt.wantErr(t, err, fmt.Sprintf("GetUploadedOrdersForUser(%v)", tt.args.order)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetUploadedOrdersForUser(%v)", tt.args.order)
			assert.IsType(t, tt.wantErrType, err)
		})
	}
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	type fields struct {
		repo *mock_dao.MockRepository
	}
	type args struct {
		order model.Order
	}
	tests := []struct {
		name          string
		args          args
		prepare       func(f *fields)
		wantErr       assert.ErrorAssertionFunc
		wantErrorType error
	}{
		{
			name: "should success update order status",
			prepare: func(f *fields) {
				id := 1
				orderInDB := &model.Order{
					ID:         &id,
					Number:     "2377225624",
					Accrual:    nil,
					Status:     "NEW",
					UploadTime: time.Unix(12345667, 3),
					User:       &model.User{ID: &id},
				}
				gomock.InOrder(
					f.repo.EXPECT().GetOrderByNumber("2377225624").Return(orderInDB, nil),
					f.repo.EXPECT().Save(orderInDB).Return(nil),
				)
			},
			args: args{order: model.Order{
				User:    nil,
				Number:  "2377225624",
				Accrual: nil,
				Status:  "INVALID",
			}},
			wantErr:       assert.NoError,
			wantErrorType: nil,
		},
		{
			name: "should return NoOrdersErr",
			prepare: func(f *fields) {
				orderInDB := &model.Order{
					ID:         nil,
					Number:     "",
					Accrual:    nil,
					Status:     "",
					UploadTime: time.Time{},
				}
				gomock.InOrder(
					f.repo.EXPECT().GetOrderByNumber("2377225624").Return(orderInDB, nil),
				)
			},
			args: args{order: model.Order{
				User:    nil,
				Number:  "2377225624",
				Accrual: nil,
				Status:  "REGISTERED",
			}},
			wantErr:       assert.Error,
			wantErrorType: &errors.NoOrdersError{},
		},
		{
			name: "should return OrderNoChangeError",
			prepare: func(f *fields) {
				id := 1
				orderInDB := &model.Order{
					ID:         &id,
					Number:     "2377225624",
					Accrual:    nil,
					Status:     "PROCESSING",
					UploadTime: time.Unix(12345667, 3),
					User:       &model.User{ID: &id},
				}
				gomock.InOrder(
					f.repo.EXPECT().GetOrderByNumber("2377225624").Return(orderInDB, nil),
				)
			},
			args: args{order: model.Order{
				User:    nil,
				Number:  "2377225624",
				Accrual: nil,
				Status:  "PROCESSING",
			}},
			wantErr:       assert.Error,
			wantErrorType: &errors.OrderNoChangeError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{repo: mock_dao.NewMockRepository(ctrl)}
			tt.prepare(&f)
			s := &OrderService{
				repo: f.repo,
			}
			orderService := s.UpdateOrderStatus(tt.args.order)
			tt.wantErr(t, orderService, fmt.Sprintf("UpdateOrderStatus(%v)", tt.args.order))
			assert.ErrorIs(t, orderService, tt.wantErrorType, fmt.Sprintf("UpdateOrderStatus(%v)", tt.args.order))
		})
	}
}

func Test_checkOrderFormat(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{number: 2377225624},
			name: "should success with right Luhn number",
			want: true,
		},
		{
			args: args{number: 1234567890},
			name: "should fail with wrong Luhn number",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkOrderFormat(tt.args.number))
		})
	}
}
