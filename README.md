# Microservices Project For Qonto

This project is a simple implementation of a microservices architecture with a producer and consumer service connected to Kafka and PostgreSQL databases.

---

## **Features**
- Producer Service:
    - Manages user accounts.
    - Sends messages to Kafka for balance initialization and transactions.
- Consumer Service:
    - Listens to Kafka for transaction and balance events.
    - Updates PostgreSQL databases for account balances.

---

## **Technologies Used**
- **Programming Language**: Go (Golang)
- **Message Broker**: Apache Kafka
- **Databases**: PostgreSQL

![Screenshot 2024-11-17 at 11 41 33 PM](https://github.com/user-attachments/assets/9e0a078d-c7e0-412c-a65d-94330cbd3d6c)
