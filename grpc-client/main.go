package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ticketpb "gRPC/proto/ticket"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := ticketpb.NewTicketServiceClient(conn)

	log.Println("🚀 Connected to Ticket gRPC Service")

	// Example 1: Create a ticket
	fmt.Println("\n=== Creating a new ticket ===")
	createResp, err := client.CreateTicket(context.Background(), &ticketpb.CreateTicketRequest{
		Title:       "Fix authentication bug",
		Description: "Users can't login with Google OAuth",
		Priority:    ticketpb.TicketPriority_TICKET_PRIORITY_HIGH,
		AssigneeId:  "user-123",
		Tags:        []string{"bug", "authentication", "urgent"},
	})
	if err != nil {
		log.Printf("Failed to create ticket: %v", err)
	} else {
		fmt.Printf("✅ Created ticket: %s - %s\n", createResp.Ticket.Id, createResp.Ticket.Title)
		ticketID := createResp.Ticket.Id

		// Example 2: Get the ticket we just created
		fmt.Println("\n=== Getting the ticket ===")
		getResp, err := client.GetTicket(context.Background(), &ticketpb.GetTicketRequest{
			Id: ticketID,
		})
		if err != nil {
			log.Printf("Failed to get ticket: %v", err)
		} else {
			ticket := getResp.Ticket
			fmt.Printf("📋 Ticket Details:\n")
			fmt.Printf("   ID: %s\n", ticket.Id)
			fmt.Printf("   Title: %s\n", ticket.Title)
			fmt.Printf("   Description: %s\n", ticket.Description)
			fmt.Printf("   Status: %s\n", ticket.Status.String())
			fmt.Printf("   Priority: %s\n", ticket.Priority.String())
			fmt.Printf("   Assignee: %s\n", ticket.AssigneeId)
			fmt.Printf("   Tags: %v\n", ticket.Tags)
		}

		// Example 3: Update the ticket
		fmt.Println("\n=== Updating the ticket ===")
		updateResp, err := client.UpdateTicket(context.Background(), &ticketpb.UpdateTicketRequest{
			Id:     ticketID,
			Status: ticketpb.TicketStatus_TICKET_STATUS_IN_PROGRESS,
			Title:  "Fix authentication bug - URGENT",
		})
		if err != nil {
			log.Printf("Failed to update ticket: %v", err)
		} else {
			fmt.Printf("✅ Updated ticket: %s - Status: %s\n",
				updateResp.Ticket.Id, updateResp.Ticket.Status.String())
		}
	}

	// Example 4: List all tickets
	fmt.Println("\n=== Listing all tickets ===")
	listResp, err := client.ListTickets(context.Background(), &ticketpb.ListTicketsRequest{
		PageSize: 10,
	})
	if err != nil {
		log.Printf("Failed to list tickets: %v", err)
	} else {
		fmt.Printf("📋 Found %d tickets:\n", len(listResp.Tickets))
		for i, ticket := range listResp.Tickets {
			fmt.Printf("   %d. %s - %s [%s]\n",
				i+1, ticket.Id[:8], ticket.Title, ticket.Status.String())
		}
	}

	// Example 5: Create another ticket
	fmt.Println("\n=== Creating another ticket ===")
	_, err = client.CreateTicket(context.Background(), &ticketpb.CreateTicketRequest{
		Title:       "Add dark mode",
		Description: "Users requested dark mode for better UX",
		Priority:    ticketpb.TicketPriority_TICKET_PRIORITY_MEDIUM,
		AssigneeId:  "user-456",
		Tags:        []string{"feature", "ui", "enhancement"},
	})
	if err != nil {
		log.Printf("Failed to create second ticket: %v", err)
	} else {
		fmt.Println("✅ Created second ticket successfully")
	}

	fmt.Println("\n🎉 gRPC Client operations completed!")
}

// Advanced example: Client with timeout and error handling
func createClientWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := ticketpb.NewTicketServiceClient(conn)

	// Create ticket with timeout
	resp, err := client.CreateTicket(ctx, &ticketpb.CreateTicketRequest{
		Title: "Timeout test ticket",
	})
	if err != nil {
		log.Printf("Error with timeout: %v", err)
		return
	}

	fmt.Printf("Created ticket with timeout: %s\n", resp.Ticket.Id)
}
