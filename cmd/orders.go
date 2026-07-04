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

type OrderStatus struct {
	Name        string `json:"name,omitempty"`
	FullName    string `json:"full_name"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Progress    int    `json:"progress,omitempty"`
}

type OrderProduct struct {
	Name    string `json:"name"`
	Carrier string `json:"carrier"`
}

type OrderUserCompany struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OrderUser struct {
	ID       string            `json:"id"`
	FullName string            `json:"full_name"`
	Email    string            `json:"email"`
	Type     string            `json:"type,omitempty"`
	Company  *OrderUserCompany `json:"company,omitempty"`
}

type OrderEvent struct {
	Type      string             `json:"type"`
	Summary   string             `json:"summary"`
	User      *OrderUser         `json:"user,omitempty"`
	Content   *OrderEventContent `json:"content,omitempty"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at,omitempty"`
}

type OrderEventContent struct {
	Status      *OrderStatus `json:"status,omitempty"`
	Body        string       `json:"body,omitempty"`
	OriginEmail string       `json:"origin_email,omitempty"`
}

type BroadbandOrderListItem struct {
	ID        string       `json:"id"`
	Reference string       `json:"reference"`
	EndUser   string       `json:"end_user,omitempty"`
	Postcode  string       `json:"postcode,omitempty"`
	DueDate   string       `json:"due_date,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at,omitempty"`
	Status    *OrderStatus `json:"status,omitempty"`
	Product   OrderProduct `json:"product"`
}

type BroadbandOrderAddress struct {
	CombinedAddress      string `json:"combined_address"`
	Telephone            string `json:"telephone,omitempty"`
	Postcode             string `json:"postcode"`
	OnsiteContact        string `json:"onsite_contact,omitempty"`
	OnsiteContactNumber  string `json:"onsite_contact_number,omitempty"`
	RequestedAppointment string `json:"requested_appointment_date,omitempty"`
}

type BroadbandOrderDetail struct {
	ID               string                `json:"id"`
	Reference        string                `json:"reference"`
	Type             string                `json:"type"`
	Carrier          string                `json:"carrier,omitempty"`
	Term             *int                  `json:"term,omitempty"`
	CustomerName     string                `json:"customer_name,omitempty"`
	Notes            string                `json:"notes,omitempty"`
	DueDate          string                `json:"due_date,omitempty"`
	CreatedAt        string                `json:"created_at"`
	CarbonCopyEmails []string              `json:"carbon_copy_emails"`
	Address          BroadbandOrderAddress `json:"address"`
	User             OrderUser             `json:"user"`
	Status           *OrderStatus          `json:"status,omitempty"`
}

type EthernetOrderListItem struct {
	ID        string       `json:"id"`
	Reference string       `json:"reference"`
	EndUser   string       `json:"end_user,omitempty"`
	Postcode  string       `json:"postcode,omitempty"`
	DueDate   string       `json:"due_date,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at,omitempty"`
	Status    *OrderStatus `json:"status,omitempty"`
	Product   OrderProduct `json:"product"`
}

type EthernetOrderAddress struct {
	Postcode            string `json:"postcode"`
	CombinedAddress     string `json:"combined_address"`
	OnSiteContact       string `json:"on_site_contact"`
	OnSiteContactNumber string `json:"on_site_contact_number"`
	Room                string `json:"room"`
	Rack                string `json:"rack"`
	Floor               string `json:"floor"`
	Notes               string `json:"notes"`
	PresentationType    string `json:"presentation_type"`
}

type EthernetOrderDetail struct {
	ID               string               `json:"id"`
	Reference        string               `json:"reference"`
	Term             int                  `json:"term"`
	Type             string               `json:"type"`
	Carrier          string               `json:"carrier"`
	CustomerName     string               `json:"customer_name"`
	PriceAmortised   bool                 `json:"price_amortised"`
	DueDate          string               `json:"due_date,omitempty"`
	Address          EthernetOrderAddress `json:"address"`
	User             OrderUser            `json:"user"`
	CarbonCopyEmails []string             `json:"carbon_copy_emails"`
	Status           *OrderStatus         `json:"status,omitempty"`
	CreatedAt        string               `json:"created_at"`
}

func ordersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orders",
		Short: "View provisioning orders",
	}

	cmd.AddCommand(broadbandOrdersCmd())
	cmd.AddCommand(ethernetOrdersCmd())
	return cmd
}

func broadbandOrdersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadband",
		Short: "Broadband orders",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all broadband orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBroadbandOrdersList(cmd)
		},
	}
	addListFlags(listCmd)

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get broadband order details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrderGet[BroadbandOrderDetail](cmd.Context(), "broadband", args[0])
		},
	}

	eventsCmd := &cobra.Command{
		Use:   "events <id>",
		Short: "List broadband order events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrderEvents(cmd.Context(), "broadband", args[0])
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(getCmd)
	cmd.AddCommand(eventsCmd)
	return cmd
}

func ethernetOrdersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethernet",
		Short: "Ethernet orders",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all ethernet orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEthernetOrdersList(cmd)
		},
	}
	addListFlags(listCmd)

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get ethernet order details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrderGet[EthernetOrderDetail](cmd.Context(), "ethernet", args[0])
		},
	}

	eventsCmd := &cobra.Command{
		Use:   "events <id>",
		Short: "List ethernet order events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrderEvents(cmd.Context(), "ethernet", args[0])
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(getCmd)
	cmd.AddCommand(eventsCmd)
	return cmd
}

func runBroadbandOrdersList(cmd *cobra.Command) error {
	flags := getListFlags(cmd)
	sp := ui.NewSpinner("Fetching broadband orders…")
	orders, err := fetchAll[BroadbandOrderListItem](cmd.Context(), "/api/orders/broadband", flags)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(orders)
	}

	if len(orders) == 0 {
		ui.Info("No broadband orders found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  REFERENCE\tEND USER\tPRODUCT\tSTATUS\tCREATED\n")
	for _, o := range orders {
		status := ""
		if o.Status != nil {
			status = o.Status.FullName
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n", o.Reference, o.EndUser, o.Product.Name, status, o.CreatedAt)
	}
	_ = w.Flush()
	fmt.Printf("\n%d order(s) found.\n", len(orders))
	return nil
}

func runEthernetOrdersList(cmd *cobra.Command) error {
	flags := getListFlags(cmd)
	sp := ui.NewSpinner("Fetching ethernet orders…")
	orders, err := fetchAll[EthernetOrderListItem](cmd.Context(), "/api/orders/ethernet", flags)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(orders)
	}

	if len(orders) == 0 {
		ui.Info("No ethernet orders found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  REFERENCE\tEND USER\tPRODUCT\tCARRIER\tSTATUS\tCREATED\n")
	for _, o := range orders {
		status := ""
		if o.Status != nil {
			status = o.Status.FullName
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%s\n", o.Reference, o.EndUser, o.Product.Name, o.Product.Carrier, status, o.CreatedAt)
	}
	_ = w.Flush()
	fmt.Printf("\n%d order(s) found.\n", len(orders))
	return nil
}

func runOrderGet[T any](ctx context.Context, orderType, id string) error {
	path := fmt.Sprintf("/api/orders/%s/%s", orderType, id)
	var order T
	sp := ui.NewSpinner("Fetching order…")
	err := authedGet(ctx, path, &order)
	sp.Stop()
	if err != nil {
		return err
	}
	return printJSON(order)
}

func runOrderEvents(ctx context.Context, orderType, id string) error {
	path := fmt.Sprintf("/api/orders/%s/%s/events", orderType, id)
	var events []OrderEvent
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
