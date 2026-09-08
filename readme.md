# msgqueue-luke

A Go service that implments RabbitMQ stable version [amqp091-go](https://github.com/rabbitmq/amqp091-go).

## Requirements

- Go 1.26 >
- Docker & Docker Compose (for running RabbitMQ locally)

## Setup

### 1. Clone & install dependencies

```bash
git clone <your-repo-url>
cd msgqueue-luke
go mod tidy
```

### 2. Configure environment variables

Create a `.env` file in the project root:

```dotenv
RABBITMQ_DIAL=amqp://lukerbtmq:lukerbtmq123@rabbitmq:5672/
```

> Use `amqp://` (or `amqps://` for TLS) followed by `user:password@host:port/vhost`.
> If you're running RabbitMQ via Docker Compose on the same network, `rabbitmq` can be used as the hostname. If running locally without Docker, use `localhost` instead.

### 3. Run RabbitMQ (Docker Compose)

```yaml
# docker-compose.yml
version: "3.8"
services:
  rabbitmq:
    image: rabbitmq:3-management #ubuntu
    container_name: rabbitmq
    ports:
      - "5672:5672"   # AMQP
      - "15672:15672" # Management UI dashboard 
    environment:
      RABBITMQ_DEFAULT_USER: urdefauluser
      RABBITMQ_DEFAULT_PASS: urpassword
```

Start it:

```bash
docker compose up -d --build
```

Check the management UI at [http://localhost:15672](http://localhost:15672) (login with the credentials above).

### 4. Run the app

```bash
go run main.go
```

## Project Structure

```
.
├── main.go
├── internals/
│   └── utils/        # config loader
├── .env
├── docker-compose.yml
└── go.mod
```

## Troubleshooting

- **`AMQP scheme must be either 'amqp://' or 'amqps://'`**
  Make sure you're passing the full connection string (e.g. `cfg.RabbitMQDial`) to `amqp091.Dial(...)`, not just the password or username.

- **Connection refused**
  Ensure the RabbitMQ container is running (`docker ps`) and that the host in your `.env` matches how you're running the app (`rabbitmq` inside Docker network, `localhost` outside it).

