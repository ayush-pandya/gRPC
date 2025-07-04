package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayush-pandya/gRPC/database"

	ticketpb "github.com/ayush-pandya/gRPC/proto/ticket"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ticketServer implements the TicketService gRPC service with PostgreSQL
type ticketServer struct {
	ticketpb.UnimplementedTicketServiceServer
	repo *database.TicketRepository
}

// newTicketServer creates a new ticket server with database repository
func newTicketServer(db *sql.DB) *ticketServer {
	return &ticketServer{
		repo: database.NewTicketRepository(db),
	}
}

// Helper functions to convert between protobuf and database models
func dbTicketToProto(dbTicket *database.Ticket) *ticketpb.Ticket {
	ticket := &ticketpb.Ticket{
		Id:        dbTicket.ID,
		Title:     dbTicket.Title,
		Status:    convertStatusToProto(dbTicket.Status),
		Priority:  convertPriorityToProto(dbTicket.Priority),
		Tags:      dbTicket.Tags,
		CreatedAt: timestamppb.New(dbTicket.CreatedAt),
		UpdatedAt: timestamppb.New(dbTicket.UpdatedAt),
	}

	if dbTicket.Description.Valid {
		ticket.Description = dbTicket.Description.String
	}

	if dbTicket.AssigneeID.Valid {
		ticket.AssigneeId = dbTicket.AssigneeID.String
	}

	return ticket
}

func convertStatusToProto(status string) ticketpb.TicketStatus {
	switch status {
	case "OPEN":
		return ticketpb.TicketStatus_TICKET_STATUS_OPEN
	case "IN_PROGRESS":
		return ticketpb.TicketStatus_TICKET_STATUS_IN_PROGRESS
	case "RESOLVED":
		return ticketpb.TicketStatus_TICKET_STATUS_RESOLVED
	case "CLOSED":
		return ticketpb.TicketStatus_TICKET_STATUS_CLOSED
	default:
		return ticketpb.TicketStatus_TICKET_STATUS_OPEN
	}
}

func convertPriorityToProto(priority string) ticketpb.TicketPriority {
	switch priority {
	case "LOW":
		return ticketpb.TicketPriority_TICKET_PRIORITY_LOW
	case "MEDIUM":
		return ticketpb.TicketPriority_TICKET_PRIORITY_MEDIUM
	case "HIGH":
		return ticketpb.TicketPriority_TICKET_PRIORITY_HIGH
	case "CRITICAL":
		return ticketpb.TicketPriority_TICKET_PRIORITY_CRITICAL
	default:
		return ticketpb.TicketPriority_TICKET_PRIORITY_MEDIUM
	}
}

func convertStatusFromProto(status ticketpb.TicketStatus) string {
	switch status {
	case ticketpb.TicketStatus_TICKET_STATUS_OPEN:
		return "OPEN"
	case ticketpb.TicketStatus_TICKET_STATUS_IN_PROGRESS:
		return "IN_PROGRESS"
	case ticketpb.TicketStatus_TICKET_STATUS_RESOLVED:
		return "RESOLVED"
	case ticketpb.TicketStatus_TICKET_STATUS_CLOSED:
		return "CLOSED"
	default:
		return "OPEN"
	}
}

func convertPriorityFromProto(priority ticketpb.TicketPriority) string {
	switch priority {
	case ticketpb.TicketPriority_TICKET_PRIORITY_LOW:
		return "LOW"
	case ticketpb.TicketPriority_TICKET_PRIORITY_MEDIUM:
		return "MEDIUM"
	case ticketpb.TicketPriority_TICKET_PRIORITY_HIGH:
		return "HIGH"
	case ticketpb.TicketPriority_TICKET_PRIORITY_CRITICAL:
		return "CRITICAL"
	default:
		return "MEDIUM"
	}
}

// CreateTicket creates a new ticket in the database
func (s *ticketServer) CreateTicket(ctx context.Context, req *ticketpb.CreateTicketRequest) (*ticketpb.CreateTicketResponse, error) {
	log.Printf("gRPC: Creating ticket in database - Title: %s", req.Title)
	log.Println(req)
	// Create database ticket
	dbTicket := &database.Ticket{
		ID:         uuid.New().String(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Title:      req.Title,
		Status:     "OPEN",
		Priority:   convertPriorityFromProto(req.Priority),
		Tags:       req.Tags,
		ReporterID: req.ReporterId,
	}

	if req.Description != "" {
		dbTicket.Description = sql.NullString{String: req.Description, Valid: true}
	}

	if req.AssigneeId != "" {
		dbTicket.AssigneeID = sql.NullString{String: req.AssigneeId, Valid: true}
	}

	// Save to database
	createdTicket, err := s.repo.Create(ctx, dbTicket)
	if err != nil {
		log.Printf("gRPC: Error creating ticket in database: %v", err)
		return nil, err
	}

	log.Printf("gRPC: Ticket created successfully in database - ID: %s", createdTicket.ID)

	return &ticketpb.CreateTicketResponse{
		Ticket: dbTicketToProto(createdTicket),
	}, nil
}

// GetTicket retrieves a ticket from the database
func (s *ticketServer) GetTicket(ctx context.Context, req *ticketpb.GetTicketRequest) (*ticketpb.GetTicketResponse, error) {
	log.Printf("gRPC: Getting ticket from database - ID: %s", req.Id)

	ticket, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		log.Printf("gRPC: Error getting ticket from database: %v", err)
		return nil, err
	}

	log.Printf("gRPC: Ticket retrieved successfully from database - ID: %s", req.Id)

	return &ticketpb.GetTicketResponse{
		Ticket: dbTicketToProto(ticket),
	}, nil
}

// ListTickets retrieves tickets from the database with pagination
func (s *ticketServer) ListTickets(ctx context.Context, req *ticketpb.ListTicketsRequest) (*ticketpb.ListTicketsResponse, error) {
	log.Println("gRPC: Listing tickets from database")

	limit := int(req.PageSize)
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	offset := 0
	// For simplicity, we'll ignore page token and just use limit
	// In production, you'd implement proper pagination with tokens

	tickets, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		log.Printf("gRPC: Error listing tickets from database: %v", err)
		return nil, err
	}

	// Convert to protobuf
	protoTickets := make([]*ticketpb.Ticket, len(tickets))
	for i, ticket := range tickets {
		protoTickets[i] = dbTicketToProto(ticket)
	}

	log.Printf("gRPC: Listed %d tickets from database", len(tickets))

	return &ticketpb.ListTicketsResponse{
		Tickets: protoTickets,
	}, nil
}

// UpdateTicket updates a ticket in the database
func (s *ticketServer) UpdateTicket(ctx context.Context, req *ticketpb.UpdateTicketRequest) (*ticketpb.UpdateTicketResponse, error) {
	log.Printf("gRPC: Updating ticket in database - ID: %s", req.Id)

	// Build updates map
	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != ticketpb.TicketStatus_TICKET_STATUS_UNSPECIFIED {
		updates["status"] = convertStatusFromProto(req.Status)
	}
	if req.Priority != ticketpb.TicketPriority_TICKET_PRIORITY_UNSPECIFIED {
		updates["priority"] = convertPriorityFromProto(req.Priority)
	}
	if req.AssigneeId != "" {
		updates["assignee_id"] = req.AssigneeId
	}
	if len(req.Tags) > 0 {
		updates["tags"] = req.Tags
	}

	// Update in database
	updatedTicket, err := s.repo.Update(ctx, req.Id, updates)
	if err != nil {
		log.Printf("gRPC: Error updating ticket in database: %v", err)
		return nil, err
	}

	log.Printf("gRPC: Ticket updated successfully in database - ID: %s", req.Id)

	return &ticketpb.UpdateTicketResponse{
		Ticket: dbTicketToProto(updatedTicket),
	}, nil
}

// DeleteTicket deletes a ticket from the database
func (s *ticketServer) DeleteTicket(ctx context.Context, req *ticketpb.DeleteTicketRequest) (*ticketpb.DeleteTicketResponse, error) {
	log.Printf("gRPC: Deleting ticket from database - ID: %s", req.Id)

	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		log.Printf("gRPC: Error deleting ticket from database: %v", err)
		return &ticketpb.DeleteTicketResponse{Success: false}, nil
	}

	log.Printf("gRPC: Ticket deleted successfully from database - ID: %s", req.Id)

	return &ticketpb.DeleteTicketResponse{Success: true}, nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	log.Println("🚀 Starting Ticket gRPC Microservice with PostgreSQL...")

	// Database configuration from environment variables
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "ayushpandya"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "ticketdb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	log.Printf("🔌 Connecting to PostgreSQL at %s:%s", dbConfig.Host, dbConfig.Port)

	// Connect to database
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create TCP listener
	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// Create gRPC server
	s := grpc.NewServer()

	// Register service with database
	ticketService := newTicketServer(db)
	ticketpb.RegisterTicketServiceServer(s, ticketService)

	log.Println("✅ Ticket Service registered with PostgreSQL backend")

	// Start server in goroutine
	go func() {
		log.Printf("🌐 Ticket gRPC Microservice listening on :%s", port)
		log.Println("🎫 Ready to handle ticket operations with PostgreSQL")
		log.Println(s.GetServiceInfo())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down Ticket gRPC Microservice...")
	s.GracefulStop()
	log.Println("👋 Ticket gRPC Microservice stopped")
}
