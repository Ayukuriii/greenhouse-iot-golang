# Greenhouse IoT Backend Service

A high-performance, concurrent IoT backend service built in **Go (Golang)** using **Fiber v3**, **PostgreSQL**, and **Eclipse Mosquitto (MQTT)** for real-time sensor ingestion, greenhouse actuator/device control, and system health monitoring.

---

## 📖 Table of Contents

- [Greenhouse IoT Backend Service](#greenhouse-iot-backend-service)
  - [📖 Table of Contents](#-table-of-contents)
  - [Overview](#overview)
  - [Key Features](#key-features)
  - [Architecture \& Tech Stack](#architecture--tech-stack)
  - [Directory Structure](#directory-structure)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Environment Configuration](#environment-configuration)
    - [Option A: Running with Docker Compose (Recommended)](#option-a-running-with-docker-compose-recommended)
    - [Option B: Running Locally (Development Mode)](#option-b-running-locally-development-mode)
  - [Database Migrations](#database-migrations)
  - [API Reference \& Examples](#api-reference--examples)
    - [1. Record Sensor Data (`POST /sensor-data`)](#1-record-sensor-data-post-sensor-data)
    - [2. Send Device Command (`POST /device-control`)](#2-send-device-command-post-device-control)
    - [3. Health \& Readiness Check (`GET /status`)](#3-health--readiness-check-get-status)
  - [MQTT Message Inspection](#mqtt-message-inspection)
    - [Subscribing to MQTT Messages](#subscribing-to-mqtt-messages)
  - [Testing](#testing)
  - [Design Document](#design-document)
  - [License](#license)

---

## Overview

The **Greenhouse IoT Backend** is designed for modern smart agriculture environments. It collects high-frequency environmental metrics (such as temperature, humidity, soil moisture) from sensors, orchestrates greenhouse actuators (ventilation fans, irrigation pumps, grow lights), and tracks command execution states reliably.

---

## Key Features

- **Concurrent Ingestion & Control**: Asynchronous, non-blocking MQTT command dispatch using Go channels and background worker goroutines.
- **Layered Clean Architecture**: Strict separation of concerns across Handlers, Services, Repositories, and Models.
- **Resilient MQTT Publishing**: At-Least-Once delivery (QoS 1) with auto-reconnect and database command status tracking (`pending` ➔ `published` / `failed`).
- **Structured JSON Logging**: Standard `log/slog` for structured, observable logging.
- **Health & Readiness Diagnostics**: Comprehensive `/status` endpoint verifying both PostgreSQL and MQTT broker connectivity.
- **Graceful Shutdown**: Controlled lifecycle termination that drains background channels and closes open connections cleanly.
- **Production Ready**: Fully containerized using multi-stage Docker builds and `docker-compose`.

---

## Architecture & Tech Stack

```
[Simulated Sensor] --HTTP POST--> [/sensor-data] --> Validation --> PostgreSQL (sensor_readings)
[Operator / App]   --HTTP POST--> [/device-control] -> Validation --> PostgreSQL (pending)
                                                                    |
                                                     [Buffered Go Channel: commandCh]
                                                                    |
                                                       [MQTT Publisher Goroutine]
                                                                    |
                                                                    v
                                                       MQTT Broker (Mosquitto)
                                                       Topic: greenhouse/control/{device_id}
[Monitoring]       --HTTP GET---> [/status] ---------> Checks PostgreSQL & MQTT Health
```

| Component | Technology | Version | Purpose |
| :--- | :--- | :--- | :--- |
| **Language** | [Go](https://go.dev/) | 1.24+ | Core programming language |
| **HTTP Framework** | [Fiber v3](https://docs.gofiber.io/) | v3.5.0 | High-performance Web/REST engine |
| **Database** | [PostgreSQL](https://www.postgresql.org/) | 16-alpine | Relational persistence & indexing |
| **DB Driver & Tool** | [pgx/v5](https://github.com/jackc/pgx) & [sqlx](https://github.com/jmoiron/sqlx) | v5 / v1.4 | High-performance SQL query mapping |
| **Message Broker** | [Eclipse Mosquitto](https://mosquitto.org/) | 2.0 | Lightweight MQTT broker |
| **MQTT Client** | [Paho MQTT Golang](https://github.com/eclipse/paho.mqtt.golang) | v1.5.0 | Go MQTT 3.1.1/5.0 client library |
| **Validation** | [validator/v10](https://github.com/go-playground/validator) | v10.25.0 | Request payload schema validation |
| **Testing** | [Testify](https://github.com/stretchr/testify) | v1.10.0 | Unit tests and repository mock suites |

---

## Directory Structure

```
greenhouse-iot-golang/
├── cmd/
│   └── api/
│       └── main.go                 # Application entrypoint & dependency injection wiring
├── internal/
│   ├── config/                     # Environment configuration loader
│   │   └── config.go
│   ├── database/                   # PostgreSQL connection & pool manager
│   │   └── postgres.go
│   ├── dto/                        # Request & response data transfer objects
│   │   ├── device_dto.go
│   │   └── sensor_dto.go
│   ├── handler/                    # Fiber v3 HTTP request controllers
│   │   ├── device_handler.go
│   │   ├── sensor_handler.go
│   │   └── status_handler.go
│   ├── middleware/                 # Standard JSON envelopes & error recovery
│   │   ├── error_handler.go
│   │   └── response.go
│   ├── model/                      # Domain entities & database models
│   │   ├── device_command.go
│   │   └── sensor.go
│   ├── mqtt/                       # MQTT client wrapper & background worker goroutine
│   │   ├── client.go
│   │   └── worker.go
│   ├── repository/                 # Database access interfaces & implementations
│   │   ├── device_repository.go
│   │   ├── device_repository_impl.go
│   │   ├── sensor_repository.go
│   │   └── sensor_repository_impl.go
│   └── service/                    # Business logic interfaces & implementations
│       ├── device_service.go
│       ├── device_service_impl.go
│       ├── sensor_service.go
│       └── sensor_service_impl.go
├── migrations/                     # Raw SQL migration files (up/down)
│   ├── 000001_create_sensor_readings.up.sql
│   ├── 000001_create_sensor_readings.down.sql
│   ├── 000002_create_device_commands.up.sql
│   └── 000002_create_device_commands.down.sql
├── mosquitto/
│   └── config/
│       └── mosquitto.conf          # Mosquitto broker configuration
├── tests/                          # Automated unit and mock tests
│   ├── device_test.go
│   └── sensor_test.go
├── .env.example                    # Sample environment variables
├── docker-compose.yml              # Complete multi-service orchestration
├── Dockerfile                      # Multi-stage production container build
├── DESIGN.md                       # Architectural decisions & system design doc
├── README.md                       # Project manual & setup guide
└── go.mod
```

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.24 or higher
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [golang-migrate](https://github.com/golang-migrate/migrate) (optional, if applying migrations manually)
- [mosquitto-clients](https://mosquitto.org/) (optional, for CLI MQTT subscription)

---

### Environment Configuration

Create a `.env` file from `.env.example`:

```bash
cp .env.example .env
```

Default variables in `.env`:

```ini
# Server Configuration
PORT=8080
ENV=development

# Database Configuration (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=greenhouse
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME_MINUTES=5

# MQTT Configuration (Eclipse Mosquitto)
MQTT_BROKER_URL=tcp://localhost:1883
MQTT_CLIENT_ID=greenhouse-backend-dev
MQTT_KEEP_ALIVE_SECONDS=60
MQTT_COMMAND_TOPIC_PREFIX=greenhouse/control

# Concurrency Channel Buffer Size
COMMAND_CHANNEL_BUFFER_SIZE=100
```

---

### Option A: Running with Docker Compose (Recommended)

Start the entire system (PostgreSQL, Mosquitto broker, and Greenhouse API service) in one command:

```bash
docker compose up --build -d
```

Check the status of running containers:

```bash
docker compose ps
```

View application logs in real time:

```bash
docker compose logs -f app
```

---

### Option B: Running Locally (Development Mode)

1. **Start PostgreSQL and Mosquitto in Docker:**
   ```bash
   docker compose up -d postgres mosquitto
   ```

2. **Apply Database Migrations:**
   ```bash
   # Using golang-migrate CLI
   migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/greenhouse?sslmode=disable" up
   ```
   *(Or run the SQL scripts in `migrations/` directly using `psql`)*

3. **Install Dependencies:**
   ```bash
   go mod download
   ```

4. **Start the Backend Service:**
   ```bash
   go run cmd/api/main.go
   ```

---

## Database Migrations

Database migrations are located in the `migrations/` directory:

| Migration File | Description |
| :--- | :--- |
| `000001_create_sensor_readings.up.sql` | Creates `sensor_readings` table and indexes on `public_id`, `device_id`, and `recorded_at`. |
| `000001_create_sensor_readings.down.sql` | Drops `sensor_readings` table. |
| `000002_create_device_commands.up.sql` | Creates `device_commands` table and indexes on `public_id`, `device_id`, and `status`. |
| `000002_create_device_commands.down.sql` | Drops `device_commands` table. |

To roll back migrations:
```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/greenhouse?sslmode=disable" down
```

---

## API Reference & Examples

All HTTP endpoints use a consistent JSON envelope format:

- **Success Response:**
  ```json
  {
    "success": true,
    "data": { ... }
  }
  ```
- **Error Response:**
  ```json
  {
    "success": false,
    "error": {
      "message": "Error description",
      "details": [
        { "field": "fieldName", "message": "Validation rule failed" }
      ]
    }
  }
  ```

---

### 1. Record Sensor Data (`POST /sensor-data`)

Stores high-frequency sensor readings (e.g., temperature, humidity).

- **URL:** `http://localhost:8080/sensor-data`
- **Method:** `POST`
- **Headers:** `Content-Type: application/json`

**Request Payload:**
```json
{
  "device_id": "sensor-gh-01",
  "sensor_type": "temperature",
  "value": 26.8,
  "unit": "celsius",
  "recorded_at": "2026-09-20T10:00:00Z"
}
```

*Note: `recorded_at` is optional; if omitted, current UTC server timestamp is used.*

**Example `curl` Request:**
```bash
curl -X POST http://localhost:8080/sensor-data \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "sensor-gh-01",
    "sensor_type": "temperature",
    "value": 26.8,
    "unit": "celsius"
  }'
```

**Success Response (`201 Created`):**
```json
{
  "success": true,
  "data": {
    "public_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "device_id": "sensor-gh-01",
    "sensor_type": "temperature",
    "value": 26.8,
    "unit": "celsius",
    "recorded_at": "2026-09-20T10:00:00Z",
    "created_at": "2026-09-20T10:00:01Z"
  }
}
```

---

### 2. Send Device Command (`POST /device-control`)

Dispatches an actuation command to a greenhouse device. The command is stored in PostgreSQL as `pending` and immediately enqueued to the MQTT publisher goroutine.

- **URL:** `http://localhost:8080/device-control`
- **Method:** `POST`
- **Headers:** `Content-Type: application/json`

**Request Payload:**
```json
{
  "device_id": "fan-zone-a",
  "command": "ON"
}
```

*Validation: `command` must be either `ON` or `OFF`.*

**Example `curl` Request:**
```bash
curl -X POST http://localhost:8080/device-control \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "fan-zone-a",
    "command": "ON"
  }'
```

**Success Response (`200 OK`):**
```json
{
  "success": true,
  "data": {
    "public_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "device_id": "fan-zone-a",
    "command": "ON",
    "status": "pending",
    "created_at": "2026-09-20T10:05:00Z"
  }
}
```

---

### 3. Health & Readiness Check (`GET /status`)

Returns diagnostic health information for the API service, PostgreSQL database connection, and MQTT broker connectivity.

- **URL:** `http://localhost:8080/status`
- **Method:** `GET`

**Example `curl` Request:**
```bash
curl -X GET http://localhost:8080/status
```

**Success Response (`200 OK` - Healthy):**
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "service": "up",
    "database": "connected",
    "mqtt": "connected",
    "timestamp": "2026-09-20T10:10:00Z"
  }
}
```

**Degraded Response (`200 OK` - Component Disconnected):**
```json
{
  "success": true,
  "data": {
    "status": "degraded",
    "service": "up",
    "database": "connected",
    "mqtt": "disconnected",
    "timestamp": "2026-09-20T10:10:00Z"
  }
}
```

---

## MQTT Message Inspection

When `POST /device-control` is called, the background worker publishes an MQTT message to the topic:

```
greenhouse/control/{device_id}
```

### Subscribing to MQTT Messages

You can subscribe to all greenhouse control topics using `mosquitto_sub` or any MQTT client:

```bash
mosquitto_sub -h localhost -p 1883 -t "greenhouse/control/#" -v
```

**Published Payload Format:**
```json
{
  "device_id": "fan-zone-a",
  "command": "ON",
  "timestamp": "2026-09-20T10:05:00Z"
}
```

---

## Testing

The project includes unit and mock test suites using `testify`:

- **Run all tests:**
  ```bash
  go test -v ./tests/...
  ```

- **Run tests with code coverage:**
  ```bash
  go test -v -cover ./tests/...
  ```

---

## Design Document

For in-depth architectural considerations, concurrency models, database indexing strategies, and trade-off analyses, refer to [DESIGN.md](DESIGN.md).

---

## License

This project is licensed under the [MIT License](LICENSE).
