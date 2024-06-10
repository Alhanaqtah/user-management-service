# User Management Service

This is a User Management microservice built using Go. It provides endpoints for user authentication, user management, and health checks. The service leverages several technologies and packages to ensure scalability, reliability, and maintainability.

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Endpoints](#endpoints)
- [Deployment with Docker Compose](#deployment-with-docker-compose)
- [License](#license)

## Features

- User authentication and authorization
- User management (create, update, delete)
- Health check endpoint
- JWT-based authentication
- Integration with RabbitMQ for message brokering
- Redis for caching
- PostgreSQL for persistent storage

## Technologies Used

- **Go**: The primary programming language used.
- **Chi**: Lightweight, idiomatic and composable router for building Go HTTP services.
- **JWT**: JSON Web Tokens for secure authentication.
- **RabbitMQ**: Message broker for asynchronous communication.
- **Redis**: In-memory data structure store, used as a cache.
- **PostgreSQL**: SQL database for persistent storage.
- **slog**: Structured logging library for Go.
- **pgx**: PostgreSQL driver and toolkit for Go.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/your-username/user-management-service.git
   cd user-management-service
   ```
2. Install the dependencies:

   ```sh
   go mod tidy
   ```

## Configuration

The service is configured using environment variables. You can create a `.env` file in the root directory of the project to set these variables.

### Environment Variables

Create a `.env` file with the following content:

```env
# ENV
ENV=local

# STORAGE
STORAGE_USER=postgres
STORAGE_PASSWORD=postgres
STORAGE_HOST=database
STORAGE_PORT=5432
STORAGE_DB=postgres
STORAGE_SSLMODE=disable
STORAGE_URL=postgres://${STORAGE_USER}:${STORAGE_PASSWORD}@${STORAGE_HOST}:${STORAGE_PORT}/${STORAGE_DB}?sslmode=${STORAGE_SSLMODE}

# CACHE
CACHE_HOST=cache
CACHE_PORT=6379
CACHE_PASSWORD=guest
CACHE_DB=0

# BROKER
BROKER_USER=guest
BROKER_PASSWORD=guest
BROKER_PORT=5672
BROKER_URL=amqp://${BROKER_USER}:${BROKER_PASSWORD}@broker
QUEUE_NAME=reset-password-stream

# SERVER
SERVER_PORT=8080
SERVER_ADDRESS=service:${SERVER_PORT}
SERVER_TIMEOUT=10s
SERVER_IDLE_TIMEOUT=4s
SERVER_SHUTDOWN_TIMEOUT=10s

# TOKENS
JWT_TOKEN_SECRET=secret
JWT_TOKEN_TTL=24h
REFRESH_TOKEN_TTL=24h
```

## Usage

1. Start the service:

   ```sh
   go run cmd/server/main.go
   ```
2. The service will be available at `http://localhost:8080`.

## Endpoints

### Health Check

- **GET /healthcheck**
  - **Description**: Check the health of the service.
  - **Response**: `200 OK` if the service is healthy.

### Authentication

- **POST /auth/login**

  - **Description**: Log in a user and return a JWT.
  - **Request**: JSON body with `username` and `password`.
  - **Response**: `200 OK` with JWT token.
- **POST /auth/register**

  - **Description**: Register a new user.
  - **Request**: JSON body with user details.
  - **Response**: `201 Created`.

### User Management

- **GET /users/me**

  - **Description**: Retrieve the logged-in user's details.
  - **Response**: `200 OK` with user details.
- **PATCH /users/me**

  - **Description**: Update the logged-in user's details.
  - **Request**: JSON body with user fields to update.
  - **Response**: `200 OK` with updated user details.
- **DELETE /users/me**

  - **Description**: Delete the logged-in user's account.
  - **Response**: `204 No Content`.

## Deployment with Docker Compose

To deploy the User Management Service using Docker Compose, follow these steps. The service configuration relies on environment variables set in a `.env` file.

### Docker Compose Configuration

Create a `docker-compose.yml` file with the following content:

```yaml
services:
  service:
    build:
      context: .
      dockerfile: dockerfile
    container_name: service
    restart: always
    ports:
      - "8080:${SERVER_PORT}"
    env_file:
      - .env
    environment:
      # HTTP SERVER
      - SERVER_ADDRESS=service:8080
      - SERVER_TIMEOUT=10s
      - SERVER_IDLE_TIMEOUT=4s
      - SERVER_SHUTDOWN_TIMEOUT=10s
      
      # CACHE
      - CACHE_HOST=${CACHE_HOST}
      - CACHE_PORT=${CACHE_PORT}
      - CACHE_DB=${CACHE_DB}

      #BROKER
      - BROKER_URL=${BROKER_URL}
      - QUEUE_NAME=${QUEUE_NAME}

      # TOKENS
      - JWT_TOKEN_SECRET=${JWT_TOKEN_SECRET}
      - JWT_TOKEN_TTL=15m
      - REFRESH_TOKEN_TTL=24h
    depends_on:
      - broker
      - database
      - cache
    networks:
      - social-services

  database:
    image: postgres:16-alpine
    container_name: database
    restart: always
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=${STORAGE_USER}
      - POSTGRES_PASSWORD=${STORAGE_PASSWORD}
      - POSTGRES_DATABASE=${STORAGE_DB}
    networks:
      - social-services

  cache:
    image: redis:7.2
    container_name: cache
    restart: always
    volumes:
        - redis_data:/data
    environment:
      - ALLOW_EMPTY_PASSWORD=yes 
    networks:
      - social-services

  broker:
    image: rabbitmq:3.13-management
    hostname: broker
    container_name: broker
    restart: always
    ports:
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    environment:
      - RABBITMQ_DEFAULT_USER=${BROKER_USER}
      - RABBITMQ_DEFAULT_PASS=${BROKER_PASSWORD}
      - RABBITMQ_SERVER_ADDITIONAL_ERL_ARGS=-rabbit disk_free_limit 2147483648
    networks:
      - social-services

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:

networks:
  social-services:
    name: social-services
    driver: bridge
```

### Running the Service

Build and start the services using Docker Compose:

```sh
docker-compose up --build
```
