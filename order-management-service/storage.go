package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Database Schema
const dbSchema = `create table if not exists customers (
	customer_id serial primary key,
	name varchar(100),
	email varchar(100)
);

create table if not exists products (
	product_id serial primary key,
	name varchar(100),
	Price DECIMAL(10, 2)
);

create table if not exists orders (
	id serial primary key,
	customer_id INT,
	product_id INT,
	quantity INT,
	total_price DECIMAL(10,2),
	status varchar(50),
	created_at timestamp,
	updated_at timestamp,
	FOREIGN KEY (customer_id) REFERENCES customers(customer_id),
	FOREIGN KEY (product_id) REFERENCES products(product_id)
);
`

type Storage interface {
	CreateOrder(*Order) error
	CreateProduct(*Product) error
	CreateCustomer(*Customer) error

	GetOrderByID(int) (*Order, error)
	GetProductByID(int) (*Product, error)
	GetCustomerByID(id int) (*Customer, error)

	UpdateOrderStatus(string, string) error
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

func (s *PostgresStore) CreateTables() error {
	_, err := s.db.Exec(dbSchema)
	return err
}

func (s *PostgresStore) GetOrderByID(id int) (*Order, error) {
	rows, err := s.db.Query("select * from orders where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanOrderValues(rows)
	}

	return nil, fmt.Errorf("order id %d not found", id)
}

func (s *PostgresStore) GetCustomerByID(id int) (*Customer, error) {

	rows, err := s.db.Query("select * from customers where customer_id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanCustomerValues(rows)
	}

	return nil, fmt.Errorf("customer id %d not found", id)
}

func (s *PostgresStore) GetProductByID(id int) (*Product, error) {
	rows, err := s.db.Query("select * from products where product_id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanProductValues(rows)
	}

	return nil, fmt.Errorf("product id %d not found", id)
}

func (s *PostgresStore) CreateOrder(order *Order) error {
	query := `insert into orders 
	(id, customer_id, product_id, quantity, total_price, status, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Query(
		query,
		order.ID,
		order.CustomerId,
		order.ProductId,
		order.Quantity,
		order.TotalPrice,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) CreateProduct(product *Product) error {
	query := `insert into products 
	(product_id, name, price)
	values ($1, $2, $3)`

	_, err := s.db.Query(
		query,
		product.ProductId,
		product.Name,
		product.Price)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) CreateCustomer(customer *Customer) error {
	query := `insert into customers 
	(customer_id, name, email)
	values ($1, $2, $3)`

	_, err := s.db.Query(
		query,
		customer.CustomerId,
		customer.Name,
		customer.Email)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateOrderStatus(orderId, status string) error {
	query := "UPDATE orders SET status=$1, updated_at=$2 WHERE id=$3"
	_, err := s.db.Exec(query, status, time.Now().UTC(), orderId)

	if err != nil {
		return err
	}

	return nil
}

func scanOrderValues(rows *sql.Rows) (*Order, error) {
	order := new(Order)
	err := rows.Scan(
		&order.ID,
		&order.CustomerId,
		&order.ProductId,
		&order.Quantity,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt)

	return order, err
}

func scanProductValues(rows *sql.Rows) (*Product, error) {
	product := new(Product)
	err := rows.Scan(
		&product.ProductId,
		&product.Name,
		&product.Price)

	return product, err
}

func scanCustomerValues(rows *sql.Rows) (*Customer, error) {
	customer := new(Customer)
	err := rows.Scan(
		&customer.CustomerId,
		&customer.Name,
		&customer.Email)

	return customer, err
}
