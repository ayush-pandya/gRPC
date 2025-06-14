package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection creates a new PostgreSQL connection
func NewConnection(config Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ PostgreSQL connection established")
	return db, nil
}

// Ticket represents a ticket in the database
type Ticket struct {
	ID          string
	Title       string
	Description sql.NullString
	Status      string
	Priority    string
	AssigneeID  sql.NullString
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ReporterID  string
}

// TicketRepository handles ticket database operations
type TicketRepository struct {
	db *sql.DB
}

// NewTicketRepository creates a new ticket repository
func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// Create creates a new ticket
func (r *TicketRepository) Create(ctx context.Context, ticket *Ticket) (*Ticket, error) {
	query := `
		INSERT INTO tickets (id, title, description, status, priority, assignee_id, tags, created_at, updated_at, reporter_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	var createdTicket Ticket
	createdTicket.ID = ticket.ID
	createdTicket.Title = ticket.Title
	createdTicket.Description = ticket.Description
	createdTicket.Status = ticket.Status
	createdTicket.Priority = ticket.Priority
	createdTicket.AssigneeID = ticket.AssigneeID
	createdTicket.Tags = ticket.Tags
	createdTicket.CreatedAt = ticket.CreatedAt
	createdTicket.UpdatedAt = ticket.UpdatedAt
	createdTicket.ReporterID = ticket.ReporterID
	log.Println("createdTicket", createdTicket)

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(ticket.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags to JSON: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		createdTicket.ID,
		ticket.Title,
		ticket.Description,
		ticket.Status,
		ticket.Priority,
		ticket.AssigneeID,
		string(tagsJSON),
		createdTicket.CreatedAt,
		createdTicket.UpdatedAt,
		createdTicket.ReporterID,
	).Scan(&createdTicket.ID, &createdTicket.CreatedAt, &createdTicket.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return &createdTicket, nil
}

// GetByID retrieves a ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, id string) (*Ticket, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, tags, created_at, updated_at
		FROM tickets WHERE id = $1`

	var ticket Ticket
	var tagsJSON string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Description,
		&ticket.Status,
		&ticket.Priority,
		&ticket.AssigneeID,
		&tagsJSON,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ticket not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Unmarshal tags from JSON
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags from JSON: %w", err)
		}
	}

	return &ticket, nil
}

// List retrieves all tickets with pagination
func (r *TicketRepository) List(ctx context.Context, limit, offset int) ([]*Ticket, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, tags, created_at, updated_at
		FROM tickets
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		var ticket Ticket
		var tagsJSON string
		err := rows.Scan(
			&ticket.ID,
			&ticket.Title,
			&ticket.Description,
			&ticket.Status,
			&ticket.Priority,
			&ticket.AssigneeID,
			&tagsJSON,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}

		// Unmarshal tags from JSON
		if tagsJSON != "" {
			err = json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags from JSON: %w", err)
			}
		}

		tickets = append(tickets, &ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tickets: %w", err)
	}

	return tickets, nil
}

// Update updates an existing ticket
func (r *TicketRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*Ticket, error) {
	// Build dynamic query based on provided updates
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "title", "description", "status", "priority", "assignee_id":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "tags":
			setParts = append(setParts, fmt.Sprintf("tags = $%d", argIndex))
			// Convert tags to JSON
			tagsJSON, err := json.Marshal(value.([]string))
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tags to JSON: %w", err)
			}
			args = append(args, string(tagsJSON))
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id) // No updates, return existing ticket
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Simplified approach - rebuild the query properly
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query := fmt.Sprintf(`
		UPDATE tickets 
		SET %s
		WHERE id = $%d
		RETURNING id, title, description, status, priority, assignee_id, tags, created_at, updated_at`,
		setClause, argIndex)

	args = append(args, id)

	var ticket Ticket
	var tagsJSON string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Description,
		&ticket.Status,
		&ticket.Priority,
		&ticket.AssigneeID,
		&tagsJSON,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ticket not found: %s", id)
		}
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	// Unmarshal tags from JSON
	if tagsJSON != "" {
		err = json.Unmarshal([]byte(tagsJSON), &ticket.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags from JSON: %w", err)
		}
	}

	return &ticket, nil
}

// Delete deletes a ticket by ID
func (r *TicketRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tickets WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ticket not found: %s", id)
	}

	return nil
}
