# Microservices Project

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

![Screenshot 2024-11-17 at 11.37.38 PM.png](..%2F..%2F..%2F..%2F..%2Fvar%2Ffolders%2Fp1%2Fbnjzf_t97vg99xq28fng42jm0000gn%2FT%2FTemporaryItems%2FNSIRD_screencaptureui_pJwnEL%2FScreenshot%202024-11-17%20at%2011.37.38%E2%80%AFPM.png)