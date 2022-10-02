package dao

import (
	"context"
	"database/sql"
	errors2 "errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/model"
)

type QueryAble interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type PostgresRepository struct {
	Conn  *sqlx.DB
	DBURI string
	db    QueryAble
}

func NewPGRepo(dbURI string) *PostgresRepository {
	conn, err := sqlx.Connect("postgres", dbURI)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	conn.SetMaxOpenConns(100)
	conn.SetMaxIdleConns(5)

	return &PostgresRepository{
		Conn:  conn,
		DBURI: dbURI,
		db:    conn,
	}
}

func (repo *PostgresRepository) Atomic(
	ctx context.Context,
	fn func(r Repository) error,
) (err error) {
	tx, err := repo.Conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
			}
		} else {
			err = tx.Commit()
		}
	}()

	newRepo := &PostgresRepository{
		Conn: repo.Conn,
		db:   tx,
	}
	err = fn(newRepo)
	return
}

func (repo PostgresRepository) Shutdown() {
	repo.Conn.SetMaxIdleConns(-1)
	repo.Conn.Close()
}

func (repo *PostgresRepository) Migrate(path string) {
	m, err := migrate.New(
		path,
		repo.DBURI)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && !errors2.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

}

func (repo *PostgresRepository) GetWithdrawalsByUserID(userID int) ([]*model.Withdraw, error) {
	var withdrawals []*model.Withdraw
	query := `
		SELECT order_num, 
		       sum, 
		       processed_at 
		FROM withdrawals 
		WHERE user_id=$1;
	`
	rows, err := repo.db.Query(query, userID)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		log.Error(err)
		return nil, err
	}
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
		withdrawals = append(withdrawals, &withdraw)
	}
	return withdrawals, nil
}

func (repo *PostgresRepository) GetBalanceByUserID(userID int) (*model.Balance, error) {
	var balance = &model.Balance{
		User: model.User{ID: &userID},
	}

	query := `
		SELECT id, balance, spent_all_time
		FROM balance
		WHERE user_id=$1;
	`

	err := repo.db.QueryRow(query, userID).Scan(
		&balance.ID,
		&balance.Balance,
		&balance.SpentAllTime,
	)
	if err != nil && !errors2.Is(err, sql.ErrNoRows) {
		log.Error(err)
		return nil, err
	}
	return balance, nil
}

func (repo *PostgresRepository) GetOrdersByUserID(userID int) ([]model.Order, error) {
	var orders []model.Order

	query := `
		SELECT id, number, upload_time, accrual, status
		FROM orders
		WHERE user_id=$1;
	`
	result, err := repo.db.Query(query, userID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer result.Close()

	err = result.Err()

	if err != nil && !errors2.Is(err, sql.ErrNoRows) {
		log.Error(err)
		return nil, err
	}

	for result.Next() {
		order := model.Order{
			User: &model.User{ID: &userID},
		}
		err = result.Scan(
			&order.ID,
			&order.Number,
			&order.UploadTime,
			&order.Accrual,
			&order.Status)
		orders = append(orders, order)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return orders, nil
}

func (repo *PostgresRepository) GetOrdersForStatusUpdate() ([]*model.Order, error) {
	var orders []*model.Order

	query := `
		SELECT id, number, upload_time, accrual, status
		FROM orders
		WHERE status in ('NEW',
		                'PROCESSING');
	`
	result, err := repo.db.Query(query)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = result.Err()
	if err != nil {
		log.Error(err)
		return nil, err
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
		if err != nil {
			log.Error(err)
			continue
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (repo *PostgresRepository) GetOrderByNumber(orderNumber string) (*model.Order, error) {
	var (
		order  = model.Order{Number: orderNumber}
		user   = model.User{}
		userID *int
	)

	query := `
		SELECT
		    id,
		    number,
		    upload_time,
		    status,
		    accrual,
		    user_id
		FROM orders 
		WHERE number=$1;
	`
	err := repo.db.QueryRow(query, orderNumber).
		Scan(
			&order.ID,
			&order.Number,
			&order.UploadTime,
			&order.Status,
			&order.Accrual,
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

func (repo *PostgresRepository) GetUser(user *model.User) (*model.User, error) {
	var userID *int
	query := `
		SELECT id FROM users WHERE username=$1 and password=$2;
	`
	err := repo.db.
		QueryRow(query, user.Login, user.Password).
		Scan(&userID)
	if err != nil && !errors2.Is(err, sql.ErrNoRows) {
		log.Error(err)
		return user, err
	}
	user.ID = userID

	return user, nil
}

func (repo *PostgresRepository) SaveWithdraw(withdraw *model.Withdraw) error {
	query := `
		INSERT INTO withdrawals(
		                        order_num, 
		                        sum, 
		                        processed_at,
		                        user_id
		                   )
		VALUES ($1, $2, $3, $4);
	`
	_, err := repo.db.Exec(query,
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

func (repo *PostgresRepository) SaveBalance(balance *model.Balance) error {
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
	_, err := repo.db.Exec(query,
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

func (repo *PostgresRepository) SaveOrder(order *model.Order) error {
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
	_, err := repo.db.Exec(query,
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

func (repo *PostgresRepository) SaveUser(user *model.User) error {
	query := `
		INSERT INTO users(
		username,
		password
		)
		VALUES($1, $2);
	`
	_, err := repo.db.Exec(query, user.Login, user.Password)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
