package dao

import (
	_ "github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
)

type PostgresRepository struct {
	Conn string
}

func NewPGRepo(dbURI string) *PostgresRepository {
	conn, err := sqlx.Connect("postgres", dbURI)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	return &PostgresRepository{Conn: dbURI}
}

func (repo *PostgresRepository) GetWithdrawals(withdraw model.Withdraw) ([]*model.Withdraw, error) {
	return getWithdrawalsForCurrentUser(withdraw, repo.Conn)
}

func (repo *PostgresRepository) GetBalance(balance model.Balance) (*model.Balance, error) {
	b, err := getCurrentUserBalance(balance, repo.Conn)
	return b, err
}

func (repo *PostgresRepository) GetOrdersForUser(order model.Order) ([]model.Order, error) {
	userID := order.User.ID
	orders := getOrdersByUserID(*userID, repo.Conn)
	return orders, nil
}

func (repo *PostgresRepository) GetOrdersForStatusUpdate() ([]*model.Order, error) {
	orders := getOrdersForUpdate(repo.Conn)
	return orders, nil
}

func (repo *PostgresRepository) GetOrderByNumber(orderNumber string) (*model.Order, error) {
	return getOrderByNumber(orderNumber, repo.Conn)
}

func (repo *PostgresRepository) GetUser(user *model.User) (*model.User, error) {
	return getUser(user, repo.Conn)
}

func (repo *PostgresRepository) Save(obj interface{}) error {
	switch x := obj.(type) {
	case *model.User:
		return saveUser(x, repo.Conn)
	case *model.Order:
		return saveOrder(x, repo.Conn)
	case *model.Balance:
		return saveBalance(x, repo.Conn)
	case *model.Withdraw:
		return saveWithdraw(x, repo.Conn)
	default:
		return errors.UnsupportedModelError{}
	}
}

func saveWithdraw(withdraw *model.Withdraw, connect string) error {
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		INSERT INTO withdrawals(
		                        order_num, 
		                        sum, 
		                        processed_at,
		                        user_id
		                   )
		VALUES ($1, $2, $3, $4);
	`
	_, err = conn.Exec(query,
		withdraw.Order,
		withdraw.Sum,
		withdraw.ProcessedAt,
		withdraw.User.ID,
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getWithdrawalsForCurrentUser(w model.Withdraw, connect string) ([]*model.Withdraw, error) {
	var withrawals []*model.Withdraw
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		SELECT order_num, 
		       sum, 
		       processed_at 
		FROM withdrawals 
		WHERE user_id=$1;
	`
	rows, err := conn.Queryx(query, w.User.ID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	_ = rows.Err()
	for rows.Next() {
		withdraw := model.Withdraw{}
		err := rows.Scan(
			&withdraw.Order,
			&withdraw.Sum,
			&withdraw.ProcessedAt,
		)
		if err != nil {
			log.Error(err)
			continue
		}
		withrawals = append(withrawals, &withdraw)
	}
	return withrawals, nil
}

func saveBalance(balance *model.Balance, connect string) error {
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		INSERT INTO balance(
		                   user_id, 
		                   balance, 
		                   spent_all_time
		                   )
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		    SET balance=$2, 
		        spent_all_time=$3 ;
	`
	_, err = conn.Exec(query,
		balance.User.ID,
		balance.Balance,
		balance.SpentAllTime,
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getCurrentUserBalance(b model.Balance, connect string) (*model.Balance, error) {
	var balance = &b
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		SELECT balance, spent_all_time
		FROM balance
		WHERE user_id=$1;
	`
	result, err := conn.Query(query, balance.User.ID)
	_ = result.Err()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer result.Close()

	result.Next()
	result.Scan(&balance.Balance, &balance.SpentAllTime)
	return balance, nil
}

func getOrdersByUserID(userID int, connect string) []model.Order {
	var orders []model.Order
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		SELECT id, number, upload_time, accrual, status
		FROM orders
		WHERE user_id=$1;
	`
	result, err := conn.Query(query, userID)
	_ = result.Err()
	if err != nil {
		log.Error(err)
		return nil
	}
	defer result.Close()

	for result.Next() {
		order := model.Order{}
		err = result.Scan(
			&order.ID,
			&order.Number,
			&order.UploadTime,
			&order.Accrual,
			&order.Status)
		orders = append(orders, order)
		if err != nil {
			log.Error(err)
			return nil
		}
	}
	return orders
}

func getOrdersForUpdate(connect string) []*model.Order {
	var orders []*model.Order
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		SELECT id, number, upload_time, accrual, status
		FROM orders
		WHERE status in ('NEW',
		                'PROCESSING');
	`
	result, err := conn.Query(query)
	_ = result.Err()
	if err != nil {
		log.Error(err)
		return nil
	}
	defer result.Close()

	for result.Next() {
		order := model.Order{}
		result.Scan(
			&order.ID,
			&order.Number,
			&order.UploadTime,
			&order.Accrual,
			&order.Status)
		orders = append(orders, &order)
	}
	return orders
}

func saveOrder(order *model.Order, connect string) error {
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		INSERT INTO orders(
		                   user_id, 
		                   number, 
		                   status,
		                   upload_time,
		                   accrual
		                   )
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (number) DO 
		    UPDATE SET 	user_id=$1,
		            	status=$3,
		            	upload_time=$4,
		            	accrual=$5;
	`
	_, err = conn.Exec(query,
		order.User.ID,
		order.Number,
		order.Status,
		order.UploadTime,
		order.Accrual,
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func saveUser(user *model.User, connect string) error {
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		INSERT INTO users(
		username,
		password
		)
		VALUES($1, $2);
	`
	_, err = conn.Exec(query, user.Login, user.Password)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getUser(user *model.User, connect string) (*model.User, error) {
	var userID *int
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		SELECT id FROM users WHERE username=$1 and password=$2;
	`
	err = conn.
		QueryRow(query, user.Login, user.Password).
		Scan(&userID)
	if err != nil {
		log.Error(err)
		return user, err
	}
	user.ID = userID

	return user, nil
}

func getOrderByNumber(orderNum string, connect string) (*model.Order, error) {
	var (
		order  = model.Order{Number: orderNum}
		user   = model.User{}
		userID *int
	)
	conn, err := sqlx.Connect("postgres", connect)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()
	query := `
		SELECT
		    id,
		    number,
		    upload_time,
		    status,
		    user_id
		FROM orders 
		WHERE number=$1;
	`
	err = conn.QueryRow(query, orderNum).
		Scan(
			&order.ID,
			&order.Number,
			&order.UploadTime,
			&order.Status,
			&userID,
		)
	if err != nil {
		log.Error(err)
		return &order, nil
	}
	user.ID = userID
	order.User = &user
	return &order, nil
}
