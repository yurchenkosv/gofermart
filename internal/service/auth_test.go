package service

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_dao "github.com/yurchenkosv/gofermart/internal/mocks"
	"github.com/yurchenkosv/gofermart/internal/model"
	"testing"
)

func TestAuthenticateUser(t *testing.T) {
	type mockBehavior func(s *mock_dao.MockRepository, user *model.User, id int)
	type args struct {
		user *model.User
	}
	tests := []struct {
		name     string
		id       int
		args     args
		behavior mockBehavior
		want     *model.User
		wantErr  bool
	}{
		{
			args: args{user: &model.User{
				Login:    "test",
				Password: "test",
			}},
			id: 1,
			behavior: func(s *mock_dao.MockRepository, user *model.User, id int) {
				user.ID = &id
				s.EXPECT().GetUser(user).Return(user, nil)
			},
			name: "should successfully return user",
			want: &model.User{
				Login:    "test",
				Password: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authRepo := mock_dao.NewMockRepository(ctrl)
			tt.behavior(authRepo, tt.args.user, tt.id)
			authService := NewAuthService(authRepo)
			got, err := authService.AuthenticateUser(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got.ID, "user id is nil")
			assert.Equal(t, tt.want.Login, got.Login)
			assert.Equal(t, tt.want.Password, got.Password)
		})
	}
}

func TestRegisterUser(t *testing.T) {
	type mockBehavior func(s *mock_dao.MockRepository, user *model.User, id int)
	type args struct {
		user *model.User
	}
	tests := []struct {
		name     string
		id       int
		args     args
		behavior mockBehavior
		want     *model.User
		wantErr  bool
	}{
		{
			args: args{user: &model.User{
				Login:    "test",
				Password: "test",
			}},
			id: 1,
			behavior: func(s *mock_dao.MockRepository, user *model.User, id int) {
				s.EXPECT().GetUser(user).Return(user, nil)
				s.EXPECT().Save(user).Return(nil)
				s.EXPECT().GetUser(user).Return(&model.User{
					ID:       &id,
					Login:    user.Login,
					Password: user.Password,
				}, nil)
			},
			name:    "should successfully save user",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authRepo := mock_dao.NewMockRepository(ctrl)
			authService := NewAuthService(authRepo)
			tt.behavior(authRepo, tt.args.user, tt.id)
			got, err := authService.RegisterUser(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, got.ID, "user id is nil")
		})
	}
}
