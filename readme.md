# msgqueue-luke

A Go service that implments RabbitMQ stable version [amqp091-go](https://github.com/rabbitmq/amqp091-go).

## Requirements

- Go 1.26 >
- Docker & Docker Compose (for running RabbitMQ locally)
- Use in-Memory database using sync.Map package for simplicity

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
RABBITMQ_DIAL=amqp://urdefaultuser:userpassword@rabbitmq:5672/

#dont forget define this username & password in docker-compose
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

### 4. Run SeaweedFS S3

The Docker Compose configuration includes a SeaweedFS service using the
`chrislusf/seaweedfs:latest` image. Its S3-compatible API listens on port
`8333` and stores data in the `s3-data` Docker volume.

Start SeaweedFS:

```bash
docker compose up -d s3
```

Configure the AWS S3 client settings in `app.env`:

```dotenv
ACCESS_KEY_S3=<seaweedfs-access-key>
SECRET_KEY_S3=<seaweedfs-secret-key>
ENDPOINT_S3=http://localhost:8333
```

Use `http://s3:8333` for `ENDPOINT_S3` when the application runs as a Docker
Compose service; use `http://localhost:8333` when running it directly on the
host. Set the access and secret keys to match the credentials configured for
the SeaweedFS S3 API. The application uses the AWS SDK for Go v2 with this
custom endpoint, and creates the `any-name-you-want` bucket on startup if it does not exist.

For a quick connectivity check, open [http://localhost:8333](http://localhost:8333).

### 5. Run the app

```bash
go run main.go
```

## Project Structure

```
.
├── app.env
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── main.go
├── cmd/
│   ├── producer/
│   │   └── producer.go
│   └── worker/
│       └── worker.go
├── internals/
│   ├── db/
│   ├── domain/
│   ├── mail/
│   ├── msg_queue/
│   ├── router/
│   ├── service/
│   ├── storage/
│   └── utils/
├── readme.md
├── loopitem.sh
├── multicurl.sh
├── downloaded.csv
├── exmpl.app.env
├── exmpl.env
└── note.txt
```

## Troubleshooting

- **`AMQP scheme must be either 'amqp://' or 'amqps://'`**
  Make sure you're passing the full connection string (e.g. `cfg.RabbitMQDial`) to `amqp091.Dial(...)`, not just the password or username.

- **Connection refused**
  Ensure the RabbitMQ container is running (`docker ps`) and that the host in your `.env` matches how you're running the app (`rabbitmq` inside Docker network, `localhost` outside it).

