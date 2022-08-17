package dao

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yurchenkosv/gofermart/internal/model"
	"testing"
	"time"
)

func getPointerFromInt(i int) *int {
	return &i
}
func getPointerFromFloat32(i float32) *float32 {
	return &i
}
func initContainers(t *testing.T, ctx context.Context) testcontainers.Container {

	port, err := nat.NewPort("tcp", "5432")
	if err != nil {
		t.Error(err)
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:12",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "gofermart",
		},
		WaitingFor: wait.ForListeningPort(port),
	}
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	return postgres
}

func TestNewPGRepo(t *testing.T) {
	type args struct {
		dbURI string
	}
	tests := []struct {
		name   string
		args   args
		before func(t *testing.T, ctx context.Context) testcontainers.Container
		want   *PostgresRepository
	}{
		{
			name:   "should create connection to db",
			args:   args{},
			before: initContainers,
			want:   &PostgresRepository{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			postgres := tt.before(t, ctx)
			defer postgres.Terminate(ctx)
			endpoint, err := postgres.Endpoint(ctx, "")
			if err != nil {
				t.Error(err)
			}
			tt.args.dbURI = fmt.Sprintf("postgresql://postgres:postgres@%s/gofermart?sslmode=disable", endpoint)
			repo := NewPGRepo(tt.args.dbURI)
			assert.IsType(t, tt.want, repo)
		})
	}
}

func Test_getCurrentUserBalance(t *testing.T) {
	type args struct {
		b    model.Balance
		repo *PostgresRepository
		qry  string
	}
	tests := []struct {
		name    string
		args    args
		before  func(t *testing.T, ctx context.Context) testcontainers.Container
		want    *model.Balance
		wantErr bool
	}{
		{
			name: "should get user balance",
			args: args{b: model.Balance{User: model.User{ID: getPointerFromInt(1)}},
				qry: `
					INSERT INTO balance(id, user_id, balance, spent_all_time) 
					VALUES (1, 1, 200, 10);
				`,
			},
			before: initContainers,
			want: &model.Balance{
				ID:           1,
				User:         model.User{ID: getPointerFromInt(1)},
				Balance:      float32(200),
				SpentAllTime: float32(10),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			postgres := tt.before(t, ctx)
			defer postgres.Terminate(ctx)
			endpoint, err := postgres.Endpoint(ctx, "")
			if err != nil {
				t.Error(err)
			}
			tt.args.repo = NewPGRepo(fmt.Sprintf("postgresql://postgres:postgres@%s/gofermart?sslmode=disable", endpoint))
			tt.args.repo.Migrate("file://../../db/migrations")
			conn, _ := sqlx.Connect("postgres", tt.args.repo.DBURI)
			conn.MustExec(tt.args.qry)

			got, err := tt.args.repo.GetBalanceByUserID(*tt.args.b.User.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCurrentUserBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getOrderByNumber(t *testing.T) {
	type args struct {
		orderNum string
		repo     *PostgresRepository
		qry      string
	}
	tests := []struct {
		name    string
		before  func(t *testing.T, ctx context.Context) testcontainers.Container
		args    args
		want    *model.Order
		wantErr bool
	}{
		{
			name:   "should return order by order number",
			before: initContainers,
			args: args{
				orderNum: "2377225624",
				qry: `
					INSERT INTO orders(id, user_id, number, upload_time, accrual, status) 
					VALUES (1, 1, '2377225624', '2022-08-16 20:32:59.390583+03', 200, 'PROCESSED')
				`,
			},
			want: &model.Order{
				ID:      getPointerFromInt(1),
				User:    &model.User{ID: getPointerFromInt(1)},
				Number:  "2377225624",
				Accrual: getPointerFromFloat32(float32(200)),
				Status:  "PROCESSED",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			postgres := tt.before(t, ctx)
			defer postgres.Terminate(ctx)
			endpoint, err := postgres.Endpoint(ctx, "")
			if err != nil {
				t.Error(err)
			}
			tt.args.repo = NewPGRepo(fmt.Sprintf("postgresql://postgres:postgres@%s/gofermart?sslmode=disable", endpoint))
			tt.args.repo.Migrate("file://../../db/migrations")
			conn, _ := sqlx.Connect("postgres", tt.args.repo.DBURI)
			conn.MustExec(tt.args.qry)

			got, err := tt.args.repo.GetOrderByNumber(tt.args.orderNum)
			tt.want.UploadTime = got.UploadTime
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrderByNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getOrdersByUserID(t *testing.T) {
	type args struct {
		userID int
		qry    string
		repo   *PostgresRepository
	}
	tests := []struct {
		name   string
		args   args
		before func(t *testing.T, ctx context.Context) testcontainers.Container
		want   []model.Order
	}{
		{
			name:   "should return user orders list",
			before: initContainers,
			args: args{
				userID: 1,
				qry: `
					INSERT INTO orders(id, user_id, number, upload_time, accrual, status) 
					VALUES (1, 1, '2377225624', '2022-08-16 20:32:59.390583+03', 200, 'PROCESSED')
				`,
			},
			want: []model.Order{
				{
					ID:         nil,
					User:       &model.User{ID: getPointerFromInt(1)},
					Number:     "2377225624",
					Accrual:    getPointerFromFloat32(float32(200)),
					Status:     "PROCESSED",
					UploadTime: time.Time{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			postgres := tt.before(t, ctx)
			defer postgres.Terminate(ctx)
			endpoint, err := postgres.Endpoint(ctx, "")
			if err != nil {
				t.Error(err)
			}
			tt.args.repo = NewPGRepo(fmt.Sprintf("postgresql://postgres:postgres@%s/gofermart?sslmode=disable", endpoint))
			tt.args.repo.Migrate("file://../../db/migrations")
			conn, _ := sqlx.Connect("postgres", tt.args.repo.DBURI)
			conn.MustExec(tt.args.qry)

			got, err := tt.args.repo.GetOrdersByUserID(tt.args.userID)

			//Не совсем понятно как сделать дату идентичной той, которая вытаскивается с БД (
			tt.want[0].UploadTime = got[0].UploadTime
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgresRepository_GetBalanceByUserID(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Balance
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			got, err := repo.GetBalanceByUserID(tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetBalanceByUserID(%v)", tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetBalanceByUserID(%v)", tt.args.userID)
		})
	}
}

func TestPostgresRepository_GetOrderByNumber(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		orderNumber string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Order
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			got, err := repo.GetOrderByNumber(tt.args.orderNumber)
			if !tt.wantErr(t, err, fmt.Sprintf("GetOrderByNumber(%v)", tt.args.orderNumber)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetOrderByNumber(%v)", tt.args.orderNumber)
		})
	}
}

func TestPostgresRepository_GetOrdersByUserID(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.Order
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			got, err := repo.GetOrdersByUserID(tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetOrdersByUserID(%v)", tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetOrdersByUserID(%v)", tt.args.userID)
		})
	}
}

func TestPostgresRepository_GetOrdersForStatusUpdate(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*model.Order
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			got, err := repo.GetOrdersForStatusUpdate()
			if !tt.wantErr(t, err, fmt.Sprintf("GetOrdersForStatusUpdate()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetOrdersForStatusUpdate()")
		})
	}
}

func TestPostgresRepository_GetUser(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		user *model.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.User
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			got, err := repo.GetUser(tt.args.user)
			if !tt.wantErr(t, err, fmt.Sprintf("GetUser(%v)", tt.args.user)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetUser(%v)", tt.args.user)
		})
	}
}

func TestPostgresRepository_GetWithdrawalsByUserID(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		userID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Withdraw
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			got, err := repo.GetWithdrawalsByUserID(tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetWithdrawalsByUserID(%v)", tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetWithdrawalsByUserID(%v)", tt.args.userID)
		})
	}
}

func TestPostgresRepository_Migrate(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			repo.Migrate(tt.args.path)
		})
	}
}

func TestPostgresRepository_SaveBalance(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		balance *model.Balance
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			tt.wantErr(t, repo.SaveBalance(tt.args.balance), fmt.Sprintf("SaveBalance(%v)", tt.args.balance))
		})
	}
}

func TestPostgresRepository_SaveOrder(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		order *model.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			tt.wantErr(t, repo.SaveOrder(tt.args.order), fmt.Sprintf("SaveOrder(%v)", tt.args.order))
		})
	}
}

func TestPostgresRepository_SaveUser(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		user *model.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			tt.wantErr(t, repo.SaveUser(tt.args.user), fmt.Sprintf("SaveUser(%v)", tt.args.user))
		})
	}
}

func TestPostgresRepository_SaveWithdraw(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	type args struct {
		withdraw *model.Withdraw
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			tt.wantErr(t, repo.SaveWithdraw(tt.args.withdraw), fmt.Sprintf("SaveWithdraw(%v)", tt.args.withdraw))
		})
	}
}

func TestPostgresRepository_Shutdown(t *testing.T) {
	type fields struct {
		Conn  *sqlx.DB
		DbURI string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := PostgresRepository{
				Conn:  tt.fields.Conn,
				DBURI: tt.fields.DbURI,
			}
			repo.Shutdown()
		})
	}
}
