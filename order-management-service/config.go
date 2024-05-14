package main

import "os"

type ServerConfig struct {
	AmqpUrl             string
	ExchangeName        string
	OrdersQueue         string
	PaymentsStatusQueue string
	Port                string
}

type DbConfig struct {
	User     string
	Password string
	Name     string
	Host     string
}

func InitConfig() (*ServerConfig, *DbConfig) {
	return &ServerConfig{
			AmqpUrl:             os.Getenv(amqpUrlStr),
			ExchangeName:        os.Getenv(exchangeNameStr),
			OrdersQueue:         os.Getenv(sendRoutingKeyStr),
			PaymentsStatusQueue: os.Getenv(receiveRoutingKeyStr),
			Port:                os.Getenv(portStr),
		}, &DbConfig{
			User:     os.Getenv(pgUserStr),
			Password: os.Getenv(pgPassStr),
			Name:     os.Getenv(pgDbStr),
			Host:     os.Getenv(dbHostStr),
		}
}
