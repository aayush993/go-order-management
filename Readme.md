# go-order-management
This repo implements an application for order management and payment processing using golang. 


#### Design:
To implement simple order processing system. 
System consists of four components 
1. Microservice for order management. 
2. Microservice for payment processing.
3. Database: Postgres
4. Message broker for asynchronous communication between two microservices

#### Order Management Microservice
Order management service will run as a microservice in a dockerized environment.
Service capabilities: 
- Ability to create an order and save the details in database. 
- Ability to publish order to rabbitmq for asynchronous processing by payment processing microservice.
- Ability to serve two API routes: 
    1. Create order: /orders
    2. Retrieve order details: /orders/{order-id}
- Worker process to monitor responses from payment processing microservice and update order status.


#### Payment Processing Microservice
Payment processing service will run as a microservice in a dockerized environment.
Service capabilities: 
- Worker process to monitor rabbitmq for requests coming from order management microservice.
- Publish response back to rabbitmq.

#### Database
Postgres database will run in a seperate docker container.
Database configuration:
- Orders table - To track order details and status. 
- Customers table - To track customer details.
- Products table - To track product details.

Customers and Products will be seeded with one entry each by order management microservice while boot-up.
For database schema, please refer: [storage.go](https://github.com/aayush993/go-order-management/blob/master/order-management-service/storage.go)

#### Message broker 
RabbitMQ will be used as message broker. It will be deployed as seperate docker container. 
Configuration:
- Two queues working in a work queue pattern
- "processsingorders" queue for order waiting for payment processing.
- "processedorders" queue for payment processing responses.
- Using direct exchange 

Assumptions: 
- payment processing will take more time. 
- Multiple payment processing microservices can consume "processsingorders" queue.

#### Deployment 
For detailed steps on deployment. Please refer: [setup.md](https://github.com/aayush993/go-order-management/blob/master/setup.md)

#### API Documentation
Note: Swagger documentation implementation is in progress using swag.

URL: http://localhost:3000

1. Create order API
    - Route: http://localhost:3000/orders
    - Example Request:
        ```
            {
            "productId": "1",
            "quantity": 1
            }
        ```
    - Example Response: 
        ```
            {
            "id": "491413",
            "customerId": "1",
            "productId": "1",
            "quantity": 1,
            "totalPrice": 199,
            "status": "Pending",
            "createdAt": "2024-05-14T21:31:53.238438244Z",
            "updatedAt": "2024-05-14T21:31:53.238438344Z"
            }
        ```

2. Get order API
    - Route: http://localhost:3000/orders/{id}
    - URL Parameters: id
    - Example URL with query params: http://localhost:3000/orders/712882

    - Example Response: 
        ```
            {
            "id": "712882",
            "customerId": "1",
            "productId": "1",
            "quantity": 1,
            "totalPrice": 199,
            "status": "Confirmed",
            "createdAt": "2024-05-14T19:52:41.487668Z",
            "updatedAt": "2024-05-14T19:52:41.512212Z"
            }
        ```


#### Enhancements possible
- Swagger documentation can be fixed.
- Unit tests can be introduced and code can be refactored to be more testable. 
- Capability to add customers and products can be introduced.
- Structured logging can be introduced.
- Log forwarding to a monitoring tool like Elasticsearch or Grafana.

