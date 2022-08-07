package dao

import (
	"context"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/model"
)

type PostgresRepository struct {
	user     *model.User
	balance  *model.Balance
	order    *model.Order
	withdraw *model.Withdraw
	Conn     string
}

func NewPGRepo(dbURI string) *PostgresRepository {
	conn, err := pgx.Connect(context.Background(), dbURI)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())
	return &PostgresRepository{Conn: dbURI}
}

func (repo *PostgresRepository) GetWithdrawals(withdraw model.Withdraw) []model.Withdraw {
	withdwals, _ := getWithdrawalsForCurrentUser(withdraw, repo.Conn)
	return withdwals
}

func (repo *PostgresRepository) SetWithdraw(withdraw *model.Withdraw) *PostgresRepository {
	repo.withdraw = withdraw
	return repo
}

func (repo *PostgresRepository) SetBalance(balance model.Balance) *PostgresRepository {
	repo.balance = &balance
	return repo
}

func (repo *PostgresRepository) GetBalance(balance model.Balance) (*model.Balance, error) {
	b, err := getCurrentUserBalance(balance, repo.Conn)
	return b, err
}

func (repo *PostgresRepository) SetOrder(order model.Order) *PostgresRepository {
	repo.order = &order
	return repo
}

func (repo *PostgresRepository) GetOrders(order model.Order) ([]model.Order, error) {
	userID := order.User.ID
	orders := getOrdersByUserID(*userID, repo.Conn)
	return orders, nil
}

func (repo *PostgresRepository) GetOrderByNumber(orderNumber int) (*model.Order, error) {
	return getOrderByNumber(orderNumber, repo.Conn)
}

func (repo *PostgresRepository) SetUser(user *model.User) *PostgresRepository {
	repo.user = user
	return repo
}

func (repo *PostgresRepository) GetUser(user *model.User) (*model.User, error) {
	return getUser(user, repo.Conn)
}

func (repo *PostgresRepository) Save() {
	if repo.user != nil {
		saveUser(repo.user, repo.Conn)
		repo.user = nil
	}
	if repo.order != nil {
		saveOrder(repo.order, repo.Conn)
		repo.order = nil
	}
	if repo.balance != nil {
		saveBalance(repo.balance, repo.Conn)
		repo.balance = nil
	}
	if repo.withdraw != nil {
		repo.withdraw = nil
	}

}

func getWithdrawalsForCurrentUser(w model.Withdraw, conn string) ([]model.Withdraw, error) {
	var withrawals []model.Withdraw
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
		return nil, err
	}
	defer connect.Close(context.Background())
	query := `
		SELECT o.number, 
		       sum, 
		       processed_at 
		FROM withdrawals 
		    JOIN orders o ON o.id = withdrawals.order_id 
		    JOIN users u ON u.id = o.user_id WHERE user_id=$1;
	`
	result, err := connect.Query(context.Background(), query, w.User.ID)
	for result.Next() {
		order := model.Order{}
		withdraw := model.Withdraw{}
		result.Scan(
			&order.Number,
			&withdraw.Sum,
			&withdraw.ProcessedAt,
		)
		withdraw.Order = order
		withrawals = append(withrawals, withdraw)
	}
	return withrawals, nil
}

func saveBalance(balance *model.Balance, conn string) error {
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
		return err
	}
	defer connect.Close(context.Background())
	query := `
		INSERT INTO balance(
		                   user_id, 
		                   balance, 
		                   spent_all_time
		                   )
		VALUES ($1, $2, $3)
		ON CONFLICT DO UPDATE SET balance=$2; 
	`
	_, err = connect.Exec(context.Background(), query,
		balance.User.ID,
		balance.Balance,
		balance.SpentAllTime,
	)
	return nil
}

func getCurrentUserBalance(b model.Balance, conn string) (*model.Balance, error) {
	var balance = &b
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
	}
	defer connect.Close(context.Background())
	query := `
		SELECT balance, spent_all_time
		FROM balance
		WHERE user_id=$1
	`
	result, err := connect.Query(context.Background(), query, balance.User.ID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer result.Close()

	result.Next()
	result.Scan(balance.Balance, balance.SpentAllTime)
	return balance, nil
}

func getOrdersByUserID(userID int, conn string) []model.Order {
	var orders []model.Order
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
	}
	defer connect.Close(context.Background())
	query := `
		SELECT number, upload_time, accrual, status
		FROM orders
		WHERE user_id=$1
	`
	result, err := connect.Query(context.Background(), query, userID)
	if err != nil {
		log.Error(err)
		return nil
	}
	defer result.Close()

	for result.Next() {
		order := model.Order{}
		result.Scan(
			&order.Number,
			&order.UploadTime,
			&order.Accrual,
			&order.Status)
		orders = append(orders, order)
	}
	return orders
}

func saveOrder(order *model.Order, conn string) error {
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
	}
	defer connect.Close(context.Background())
	query := `
		INSERT INTO orders(
		                   user_id, 
		                   number, 
		                   status,
		                   upload_time,
		                   accrual
		                   )
		VALUES ($1, $2, $3, $4, $5); 
	`
	_, err = connect.Exec(context.Background(), query,
		order.User.ID,
		order.Number,
		order.Status,
		order.UploadTime,
		order.Accrual,
	)
	return nil
}

func saveUser(user *model.User, conn string) error {
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
	}
	defer connect.Close(context.Background())
	query := `
		INSERT INTO users(
		username,
		password
		)
		VALUES($1, $2);
	`
	_, err = connect.Exec(context.Background(), query, user.Login, user.Password)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getUser(user *model.User, conn string) (*model.User, error) {
	var userID *int
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
	}
	defer connect.Close(context.Background())
	query := `
		SELECT id FROM users WHERE username=$1 and password=$2
	`
	result, err := connect.Query(context.Background(), query, user.Login, user.Password)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer result.Close()

	result.Next()
	result.Scan(&userID)
	user.ID = userID

	return user, nil
}

func getOrderByNumber(orderNum int, conn string) (*model.Order, error) {
	var order = model.Order{Number: orderNum}
	connect, err := pgx.Connect(context.Background(), conn)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
	}
	defer connect.Close(context.Background())
	query := `
		SELECT id FROM orders WHERE number=$1
	`
	result, err := connect.Query(context.Background(), query, orderNum)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer result.Close()

	result.Next()
	result.Scan(&order.ID)

	return &order, nil
}
