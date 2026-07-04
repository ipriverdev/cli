package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ipriverdev/cli/internal/app"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

type TicketStatus struct {
	Name     string `json:"name,omitempty"`
	FullName string `json:"full_name"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
}

type TicketServiceRef struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
}

type TicketCompanyRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TicketUserRef struct {
	ID       string            `json:"id"`
	FullName string            `json:"full_name"`
	Email    string            `json:"email"`
	Company  *TicketCompanyRef `json:"company,omitempty"`
}

type TicketListItem struct {
	ID          string             `json:"id"`
	Reference   string             `json:"reference"`
	Title       string             `json:"title"`
	Escalated   bool               `json:"escalated"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	CompanyName string             `json:"company_name"`
	Status      TicketStatus       `json:"status"`
	Services    []TicketServiceRef `json:"services"`
}

type TicketDetail struct {
	ID               string             `json:"id"`
	Reference        string             `json:"reference"`
	Title            string             `json:"title"`
	Status           TicketStatus       `json:"status"`
	Email            string             `json:"email"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
	Company          TicketCompanyRef   `json:"company"`
	User             *TicketUserRef     `json:"user,omitempty"`
	Services         []TicketServiceRef `json:"services"`
	CarbonCopyEmails []string           `json:"carbon_copy_emails"`
}

type TicketEvent struct {
	Type      string              `json:"type"`
	Summary   string              `json:"summary"`
	User      *TicketEventUser    `json:"user,omitempty"`
	Content   *TicketEventContent `json:"content,omitempty"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at,omitempty"`
}

type TicketEventUser struct {
	ID       string            `json:"id"`
	FullName string            `json:"full_name"`
	Email    string            `json:"email"`
	Type     string            `json:"type,omitempty"`
	Company  *TicketCompanyRef `json:"company,omitempty"`
}

type TicketEventContent struct {
	Status      *TicketStatus `json:"status,omitempty"`
	Body        string        `json:"body,omitempty"`
	OriginEmail string        `json:"origin_email,omitempty"`
}

type CreateTicketRequest struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Type              string   `json:"type"`
	InternetServiceID string   `json:"internet_service_id,omitempty"`
	CarbonCopyEmails  []string `json:"carbon_copy_emails,omitempty"`
}

type CreateTicketResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
	URL     string `json:"url"`
}

type TicketMessageRequest struct {
	Message string `json:"message"`
}

type TicketMessageResponse struct {
	Message string `json:"message"`
}

func ticketsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tickets",
		Short: "Manage support tickets",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all support tickets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketsList(cmd)
		},
	}
	addListFlags(listCmd)

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get support ticket details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketGet(cmd.Context(), args[0])
		},
	}

	eventsCmd := &cobra.Command{
		Use:   "events <id>",
		Short: "List support ticket events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketEvents(cmd.Context(), args[0])
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a support ticket",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketCreate(cmd)
		},
	}
	createCmd.Flags().String("title", "", "Short summary of the issue (required)")
	createCmd.Flags().String("description", "", "Detailed description (required)")
	createCmd.Flags().String("type", "", "Fault type: total_loss, intermittent_connectivity, change_request, query, network (required)")
	createCmd.Flags().String("service", "", "Internet service UUID to associate")
	createCmd.Flags().StringSlice("cc", nil, "CC email addresses for notifications")
	_ = createCmd.MarkFlagRequired("title")
	_ = createCmd.MarkFlagRequired("description")
	_ = createCmd.MarkFlagRequired("type")

	replyCmd := &cobra.Command{
		Use:   "reply <id>",
		Short: "Send a reply on a support ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketReply(cmd, args[0])
		},
	}
	replyCmd.Flags().String("message", "", "Message to post (required)")
	_ = replyCmd.MarkFlagRequired("message")

	cmd.AddCommand(listCmd)
	cmd.AddCommand(getCmd)
	cmd.AddCommand(eventsCmd)
	cmd.AddCommand(createCmd)
	cmd.AddCommand(replyCmd)
	return cmd
}

func runTicketsList(cmd *cobra.Command) error {
	flags := getListFlags(cmd)
	sp := ui.NewSpinner("Fetching tickets…")
	tickets, err := fetchAll[TicketListItem](cmd.Context(), "/api/tickets", flags)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(tickets)
	}

	if len(tickets) == 0 {
		ui.Info("No tickets found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  REFERENCE\tTITLE\tSTATUS\tCOMPANY\tCREATED\n")
	for _, t := range tickets {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n", t.Reference, t.Title, t.Status.FullName, t.CompanyName, t.CreatedAt)
	}
	_ = w.Flush()
	fmt.Printf("\n%d ticket(s) found.\n", len(tickets))
	return nil
}

func runTicketGet(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/tickets/%s", id)
	var ticket TicketDetail
	sp := ui.NewSpinner("Fetching ticket…")
	err := authedGet(ctx, path, &ticket)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(ticket)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%s\n", ticket.ID)
	fmt.Fprintf(w, "Reference:\t%s\n", ticket.Reference)
	fmt.Fprintf(w, "Title:\t%s\n", ticket.Title)
	fmt.Fprintf(w, "Status:\t%s\n", ticket.Status.FullName)
	fmt.Fprintf(w, "Company:\t%s\n", ticket.Company.Name)
	if ticket.User != nil {
		fmt.Fprintf(w, "User:\t%s (%s)\n", ticket.User.FullName, ticket.User.Email)
	}
	fmt.Fprintf(w, "Created:\t%s\n", ticket.CreatedAt)
	fmt.Fprintf(w, "Updated:\t%s\n", ticket.UpdatedAt)
	if len(ticket.Services) > 0 {
		for _, s := range ticket.Services {
			fmt.Fprintf(w, "Service:\t%s (%s)\n", s.Reference, s.ID)
		}
	}
	_ = w.Flush()
	return nil
}

func runTicketEvents(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/tickets/%s/events", id)
	var events []TicketEvent
	sp := ui.NewSpinner("Fetching events…")
	err := authedGet(ctx, path, &events)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(events)
	}

	if len(events) == 0 {
		ui.Info("No events found.")
		return nil
	}

	for _, e := range events {
		fmt.Printf("  [%s] %s\n", e.CreatedAt, e.Summary)
		if e.Content != nil && e.Content.Body != "" {
			fmt.Printf("    %s\n", e.Content.Body)
		}
	}
	fmt.Printf("\n%d event(s).\n", len(events))
	return nil
}

func runTicketCreate(cmd *cobra.Command) error {
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	ticketType, _ := cmd.Flags().GetString("type")
	serviceID, _ := cmd.Flags().GetString("service")
	ccEmails, _ := cmd.Flags().GetStringSlice("cc")

	req := CreateTicketRequest{
		Title:             title,
		Description:       description,
		Type:              ticketType,
		InternetServiceID: serviceID,
		CarbonCopyEmails:  ccEmails,
	}

	sp := ui.NewSpinner("Creating ticket…")
	var resp CreateTicketResponse
	err := authedPost(cmd.Context(), "/api/tickets", req, &resp)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(resp)
	}

	ui.Success(fmt.Sprintf("Ticket created: %s", resp.ID))
	return nil
}

func runTicketReply(cmd *cobra.Command, id string) error {
	message, _ := cmd.Flags().GetString("message")

	path := fmt.Sprintf("/api/tickets/%s/messages", id)
	req := TicketMessageRequest{Message: message}

	sp := ui.NewSpinner("Sending reply…")
	var resp TicketMessageResponse
	err := authedPost(cmd.Context(), path, req, &resp)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(resp)
	}

	ui.Success("Message sent.")
	return nil
}
