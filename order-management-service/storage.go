package main

import (
	"database/sql"
	"fmt"

	"github.com/aayush993/go-order-management/common"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateOrder(*common.Order) error
	GetOrderByID(int) (*common.Order, error)
	UpdateOrder(order *common.Order) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(userName, password, dbName, host string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s", userName, password, dbName, host)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createOrdersTable()
}

func (s *PostgresStore) createOrdersTable() error {
	query := `create table if not exists orders (
		id serial primary key,
		customer_name varchar(100),
		product_name varchar(100),
		quantity serial,
		amount serial,
		status varchar(100),
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) GetOrderByID(id int) (*common.Order, error) {
	rows, err := s.db.Query("select * from orders where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanOrderValues(rows)
	}

	return nil, fmt.Errorf("order %d not found", id)
}

func (s *PostgresStore) CreateOrder(order *common.Order) error {
	query := `insert into orders 
	(id, customer_name, product_name, quantity, amount, status, created_at)
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.Query(
		query,
		order.ID,
		order.CustomerName,
		order.Product,
		order.Quantity,
		order.Amount,
		order.Status,
		order.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateOrder(order *common.Order) error {
	query := "UPDATE orders SET status=$1 WHERE id=$2"
	_, err := s.db.Exec(query, order.Status, order.ID)

	if err != nil {
		return err
	}

	return nil
}

func scanOrderValues(rows *sql.Rows) (*common.Order, error) {
	order := new(common.Order)
	err := rows.Scan(
		&order.ID,
		&order.CustomerName,
		&order.Product,
		&order.Quantity,
		&order.Amount,
		&order.Status,
		&order.CreatedAt)

	return order, err
}
