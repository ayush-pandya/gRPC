# gRPC Ticket Management System

A distributed ticket management system built with gRPC, Go, and PostgreSQL. This project demonstrates a complete microservices architecture with protocol buffers for service definitions, Docker containerization, and database persistence.

## 🏗️ Architecture

```
┌─────────────────┐    gRPC     ┌─────────────────┐    SQL     ┌─────────────────┐
│   gRPC Client   │ ◄────────► │  Ticket Service │ ◄───────► │   PostgreSQL    │
│    (Port 50052) │             │   (Port 50051)  │            │   (Port 5432)   │
└─────────────────┘             └─────────────────┘            └─────────────────┘
```

### Components

- **gRPC Server** (`ticket-service-db/`): Core ticket management service with PostgreSQL integration
- **gRPC Client** (`grpc-client/`): Example client demonstrating API usage
- **PostgreSQL Database**: Persistent storage for tickets with proper indexing
- **Protocol Buffers** (`proto/`): Service definitions and data contracts

## 🚀 Features

- **CRUD Operations**: Create, Read, Update, Delete tickets
- **Ticket Management**: 
  - Multiple status levels (Open, In Progress, Resolved, Closed)
  - Priority levels (Low, Medium, High, Critical)
  - Assignee and reporter tracking
  - Tagging system with JSON support
- **Database Integration**: PostgreSQL with optimized indexes
- **Containerization**: Full Docker Compose setup
- **Type Safety**: Protocol Buffers for strongly-typed API contracts

## 🛠️ Technology Stack

- **Language**: Go 1.23.2
- **RPC Framework**: gRPC with Protocol Buffers
- **Database**: PostgreSQL 15
- **Containerization**: Docker & Docker Compose
- **Key Dependencies**:
  - `google.golang.org/grpc` - gRPC framework
  - `google.golang.org/protobuf` - Protocol Buffers
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/google/uuid` - UUID generation

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)
- Make (optional, for development scripts)

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd gRPC
   ```

2. **Start all services**:
   ```bash
   docker-compose up --build
   ```

3. **Verify services are running**:
   - gRPC Server: `localhost:50051`
   - gRPC Client: `localhost:50052`
   - PostgreSQL: `localhost:5432`

### Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Start PostgreSQL** (using Docker):
   ```bash
   docker-compose up ticket_db -d
   ```

3. **Set environment variables**:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=ayushpandya
   export DB_PASSWORD=postgres
   export DB_NAME=ticketdb
   export DB_SSLMODE=disable
   ```

4. **Run the server**:
   ```bash
   go run ticket-service-db/main.go
   ```

5. **Run the client** (in another terminal):
   ```bash
   go run grpc-client/main.go
   ```

## 📡 API Reference

### Service Definition

The `TicketService` provides the following RPC methods:

```protobuf
service TicketService {
  rpc CreateTicket(CreateTicketRequest) returns (CreateTicketResponse);
  rpc GetTicket(GetTicketRequest) returns (GetTicketResponse);
  rpc ListTickets(ListTicketsRequest) returns (ListTicketsResponse);
  rpc UpdateTicket(UpdateTicketRequest) returns (UpdateTicketResponse);
  rpc DeleteTicket(DeleteTicketRequest) returns (DeleteTicketResponse);
}
```

### Ticket Model

```protobuf
message Ticket {
  string id = 1;
  string title = 2;
  string description = 3;
  TicketStatus status = 4;
  TicketPriority priority = 5;
  string assignee_id = 6;
  repeated string tags = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}
```

### Enums

**TicketStatus**: `OPEN`, `IN_PROGRESS`, `RESOLVED`, `CLOSED`

**TicketPriority**: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`

## 🗄️ Database Schema

```sql
CREATE TABLE tickets (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    assignee_id VARCHAR(255),
    tags JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reporter_id VARCHAR(255) NOT NULL
);
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `ticket_db` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `ayushpandya` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `ticketdb` |
| `DB_SSLMODE` | SSL mode | `disable` |

### Docker Compose Services

- **grpc-server**: Ticket service (port 50051)
- **grpc-client**: Example client (port 50052)
- **ticket_db**: PostgreSQL database (port 5432)

## 🧪 Testing

### Using the Client

The included gRPC client demonstrates all available operations:

```bash
# Run the client to see example operations
docker-compose up grpc-client
```

### Manual Testing with grpcurl

```bash
# List all tickets
grpcurl -plaintext localhost:50051 ticket.TicketService/ListTickets

# Get a specific ticket
grpcurl -plaintext -d '{"id": "ticket-id"}' localhost:50051 ticket.TicketService/GetTicket
```

## 📁 Project Structure

```
gRPC/
├── docker-compose.yaml          # Docker Compose configuration
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── init.sql                     # Database initialization script
├── proto/
│   └── ticket.proto            # Protocol Buffer definitions
├── ticket-service-db/
│   ├── Dockerfile              # Server container configuration
│   └── main.go                 # gRPC server implementation
├── grpc-client/
│   ├── Dockerfile              # Client container configuration
│   └── main.go                 # Example gRPC client
└── database/
    └── postgres.go             # Database repository layer
```

## 🔄 Development Workflow

### Regenerating Protocol Buffers

After modifying `proto/ticket.proto`:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/ticket.proto
```

### Building and Running

```bash
# Build all containers
docker-compose build

# Run in development mode
docker-compose up

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f grpc-server
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Compose Reference](https://docs.docker.com/compose/)