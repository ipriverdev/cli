package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ipriverdev/cli/internal/ui"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func mcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server for Claude Desktop and other AI assistants",
		Long: `Start the built-in MCP (Model Context Protocol) server over stdio.

When run without a subcommand, starts the MCP server. This is what
Claude Desktop calls — you don't normally run this yourself.

Use "ipriver mcp install" to set up Claude Desktop automatically.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCP()
		},
	}

	cmd.AddCommand(mcpInstallCmd())
	cmd.AddCommand(mcpUninstallCmd())

	return cmd
}

func mcpInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Set up the IP River MCP server in Claude Desktop",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPInstall()
		},
	}
}

func mcpUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the IP River MCP server from Claude Desktop",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPUninstall()
		},
	}
}

func claudeConfigPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	}
}

func ipriverBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "ipriver"
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	return resolved
}

func runMCPInstall() error {
	configPath, err := claudeConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine Claude Desktop config path: %w", err)
	}

	config := make(map[string]any)

	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("could not parse %s: %w", configPath, err)
		}
	}

	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		servers = make(map[string]any)
	}

	if _, exists := servers["ipriver"]; exists {
		ui.Success("IP River is already configured in Claude Desktop.")
		ui.Info("Config: " + configPath)
		return nil
	}

	binPath := ipriverBinaryPath()

	servers["ipriver"] = map[string]any{
		"command": binPath,
		"args":    []string{"mcp"},
	}
	config["mcpServers"] = servers

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, append(out, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	ui.Success("IP River MCP server added to Claude Desktop.")
	ui.Info("Config: " + configPath)
	ui.Info("Binary: " + binPath)
	fmt.Println()
	ui.Info("Restart Claude Desktop for changes to take effect.")

	return nil
}

func runMCPUninstall() error {
	configPath, err := claudeConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine Claude Desktop config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ui.Info("No Claude Desktop config found. Nothing to remove.")
			return nil
		}
		return fmt.Errorf("read config %s: %w", configPath, err)
	}

	config := make(map[string]any)
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("could not parse %s: %w", configPath, err)
	}

	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		ui.Info("No MCP servers configured. Nothing to remove.")
		return nil
	}

	if _, exists := servers["ipriver"]; !exists {
		ui.Info("IP River is not configured in Claude Desktop. Nothing to remove.")
		return nil
	}

	delete(servers, "ipriver")
	if len(servers) == 0 {
		delete(config, "mcpServers")
	} else {
		config["mcpServers"] = servers
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, append(out, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	ui.Success("IP River MCP server removed from Claude Desktop.")
	ui.Info("Restart Claude Desktop for changes to take effect.")

	return nil
}

func runMCP() error {
	s := server.NewMCPServer(
		"ipriver",
		version,
		server.WithToolCapabilities(false),
	)

	registerAddressTools(s)
	registerCheckTools(s)
	registerServicesTools(s)
	registerOrdersTools(s)
	registerTicketsTools(s)
	registerCatalogueTools(s)
	registerAccountTools(s)

	return server.ServeStdio(s)
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal result: %s", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func mcpError(msg string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(msg), nil
}

// --- Address tools ---

func registerAddressTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("address_postcode",
			mcp.WithDescription("Search addresses by UK postcode. Returns matching addresses with UPRN, ALK, and district code."),
			mcp.WithString("postcode", mcp.Required(), mcp.Description("UK postcode to search")),
		),
		handleAddressPostcode,
	)

	s.AddTool(
		mcp.NewTool("address_uprn",
			mcp.WithDescription("Look up a single address by its Unique Property Reference Number (UPRN)."),
			mcp.WithString("uprn", mcp.Required(), mcp.Description("Unique Property Reference Number")),
		),
		handleAddressUPRN,
	)
}

func handleAddressPostcode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	postcode, err := requiredString(request, "postcode")
	if err != nil {
		return mcpError(err.Error())
	}

	var addresses []Address
	path := fmt.Sprintf("/api/address/postcode/%s", url.PathEscape(postcode))
	if err := authedGet(ctx, path, &addresses); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(addresses)
}

func handleAddressUPRN(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uprn, err := requiredString(request, "uprn")
	if err != nil {
		return mcpError(err.Error())
	}

	var addr Address
	path := fmt.Sprintf("/api/address/uprn/%s", url.PathEscape(uprn))
	if err := authedGet(ctx, path, &addr); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(addr)
}

// --- Check (availability) tools ---

func registerCheckTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("check_availability",
			mcp.WithDescription("Run a product availability check for a UK address. Submits the check, polls until complete (up to timeout), and returns available broadband/ethernet products with pricing. At least postcode or UPRN is required."),
			mcp.WithString("postcode", mcp.Description("UK postcode")),
			mcp.WithString("uprn", mcp.Description("Unique Property Reference Number")),
			mcp.WithString("alk", mcp.Description("Address Location Key (pair with dc)")),
			mcp.WithString("dc", mcp.Description("District Code (pair with alk)")),
			mcp.WithNumber("lat", mcp.Description("Latitude")),
			mcp.WithNumber("lon", mcp.Description("Longitude")),
			mcp.WithString("address", mcp.Description("Full address string")),
			mcp.WithNumber("timeout", mcp.Description("Max seconds to wait for completion (default 120)")),
		),
		handleCheckAvailability,
	)

	s.AddTool(
		mcp.NewTool("check_status",
			mcp.WithDescription("Get the current status of an availability check by its ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Availability check ID")),
		),
		handleCheckStatus,
	)

	s.AddTool(
		mcp.NewTool("check_get",
			mcp.WithDescription("Get full details of an availability check request including search parameters and user info."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Availability check ID")),
		),
		handleCheckGet,
	)

	s.AddTool(
		mcp.NewTool("check_products",
			mcp.WithDescription("Get the list of available products for a completed availability check."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Availability check ID")),
		),
		handleCheckProducts,
	)
}

func handleCheckAvailability(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	postcode := optionalString(request, "postcode")
	uprn := optionalString(request, "uprn")

	if postcode == "" && uprn == "" {
		return mcpError("at least postcode or uprn is required")
	}

	req := AvailabilityCheckRequest{
		Postcode: postcode,
		UPRN:     uprn,
		ALK:      optionalString(request, "alk"),
		DC:       optionalString(request, "dc"),
		Address:  optionalString(request, "address"),
	}

	if lat, ok := optionalFloat(request, "lat"); ok {
		req.Latitude = &lat
	}
	if lon, ok := optionalFloat(request, "lon"); ok {
		req.Longitude = &lon
	}

	timeout := 120
	if t, ok := optionalFloat(request, "timeout"); ok && t > 0 {
		timeout = int(t)
	}

	var resp AvailabilityCheckResponse
	status, err := authedPostStatus(ctx, "/api/availability-checker", req, &resp)
	if err != nil {
		return mcpError(err.Error())
	}
	if status != http.StatusAccepted {
		return mcpError(fmt.Sprintf("unexpected status %d", status))
	}

	products, finalStatus, err := mcpPollAndFetchProducts(ctx, resp.ID, timeout)
	if err != nil {
		return mcpError(err.Error())
	}

	return jsonResult(CheckResult{
		CheckID:  resp.ID,
		Status:   finalStatus,
		Products: products,
	})
}

func mcpPollAndFetchProducts(ctx context.Context, checkID string, timeoutSecs int) ([]AvailableProduct, string, error) {
	deadline := time.Now().Add(time.Duration(timeoutSecs) * time.Second)
	statusPath := fmt.Sprintf("/api/availability-checker/%s/status", checkID)
	retryInterval := 5 * time.Second

	for {
		if time.Now().After(deadline) {
			return nil, "TIMED_OUT", fmt.Errorf("timed out after %ds waiting for check %s", timeoutSecs, checkID)
		}

		var statusResp AvailabilityCheckStatus
		httpStatus, err := authedGetWithStatus(ctx, statusPath, &statusResp)
		if err != nil {
			return nil, "", err
		}

		switch {
		case httpStatus == http.StatusOK && statusResp.Status == "COMPLETED":
			return fetchProducts(ctx, checkID)
		case httpStatus == http.StatusGatewayTimeout || statusResp.Status == "TIMED_OUT":
			products, status, _ := fetchProducts(ctx, checkID)
			if status == "" {
				status = "TIMED_OUT"
			}
			return products, status, nil
		}

		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}

func handleCheckStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/availability-checker/%s/status", id)
	var resp AvailabilityCheckStatus
	if _, err := authedGetWithStatus(ctx, path, &resp); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(resp)
}

func handleCheckGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/availability-checker/%s", id)
	var resp AvailabilityCheckDetail
	if err := authedGet(ctx, path, &resp); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(resp)
}

func handleCheckProducts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/availability-checker/%s/products", id)
	var products []AvailableProduct
	if err := authedGet(ctx, path, &products); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(products)
}

// --- Services tools ---

func registerServicesTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("services_list",
			mcp.WithDescription("List internet services. Supports filtering by status, technology, network type, date range, and free-text search."),
			mcp.WithString("status", mcp.Description("Filter by status (e.g. active, ceased)")),
			mcp.WithString("technology", mcp.Description("Filter by technology (ethernet, broadband, p2p)")),
			mcp.WithString("network_type", mcp.Description("Filter by network type (on_net, off_net)")),
			mcp.WithString("search", mcp.Description("Free-text search")),
			mcp.WithString("from", mcp.Description("From date (YYYY-MM-DD)")),
			mcp.WithString("to", mcp.Description("To date (YYYY-MM-DD)")),
			mcp.WithString("sort", mcp.Description("Field to sort by")),
			mcp.WithString("direction", mcp.Description("Sort direction: ASC or DESC")),
			mcp.WithNumber("limit", mcp.Description("Max results (0 = no limit)")),
		),
		handleServicesList,
	)

	s.AddTool(
		mcp.NewTool("services_get",
			mcp.WithDescription("Get details of a single internet service by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Internet service ID")),
		),
		handleServicesGet,
	)
}

func handleServicesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flags := listFlagsFromRequest(request)
	technology := optionalString(request, "technology")
	networkType := optionalString(request, "network_type")

	services, err := fetchAll[InternetService](ctx, "/api/internet-services", flags, func(q url.Values) {
		if technology != "" {
			q.Add("technology[]", technology)
		}
		if networkType != "" {
			q.Add("networkType[]", networkType)
		}
	})
	if err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(services)
}

func handleServicesGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/internet-services/%s", id)
	var svc InternetService
	if err := authedGet(ctx, path, &svc); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(svc)
}

// --- Orders tools ---

func registerOrdersTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("orders_broadband_list",
			mcp.WithDescription("List broadband provisioning orders. Supports filtering by status, date range, and search."),
			mcp.WithString("status", mcp.Description("Filter by status")),
			mcp.WithString("search", mcp.Description("Free-text search")),
			mcp.WithString("from", mcp.Description("From date (YYYY-MM-DD)")),
			mcp.WithString("to", mcp.Description("To date (YYYY-MM-DD)")),
			mcp.WithString("sort", mcp.Description("Field to sort by")),
			mcp.WithString("direction", mcp.Description("Sort direction: ASC or DESC")),
			mcp.WithNumber("limit", mcp.Description("Max results (0 = no limit)")),
		),
		handleBroadbandOrdersList,
	)

	s.AddTool(
		mcp.NewTool("orders_broadband_get",
			mcp.WithDescription("Get details of a single broadband order."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Broadband order ID")),
		),
		handleBroadbandOrderGet,
	)

	s.AddTool(
		mcp.NewTool("orders_broadband_events",
			mcp.WithDescription("List events/timeline for a broadband order."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Broadband order ID")),
		),
		handleBroadbandOrderEvents,
	)

	s.AddTool(
		mcp.NewTool("orders_ethernet_list",
			mcp.WithDescription("List ethernet provisioning orders. Supports filtering by status, date range, and search."),
			mcp.WithString("status", mcp.Description("Filter by status")),
			mcp.WithString("search", mcp.Description("Free-text search")),
			mcp.WithString("from", mcp.Description("From date (YYYY-MM-DD)")),
			mcp.WithString("to", mcp.Description("To date (YYYY-MM-DD)")),
			mcp.WithString("sort", mcp.Description("Field to sort by")),
			mcp.WithString("direction", mcp.Description("Sort direction: ASC or DESC")),
			mcp.WithNumber("limit", mcp.Description("Max results (0 = no limit)")),
		),
		handleEthernetOrdersList,
	)

	s.AddTool(
		mcp.NewTool("orders_ethernet_get",
			mcp.WithDescription("Get details of a single ethernet order."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Ethernet order ID")),
		),
		handleEthernetOrderGet,
	)

	s.AddTool(
		mcp.NewTool("orders_ethernet_events",
			mcp.WithDescription("List events/timeline for an ethernet order."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Ethernet order ID")),
		),
		handleEthernetOrderEvents,
	)
}

func handleBroadbandOrdersList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flags := listFlagsFromRequest(request)
	orders, err := fetchAll[BroadbandOrderListItem](ctx, "/api/orders/broadband", flags)
	if err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(orders)
}

func handleBroadbandOrderGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/orders/broadband/%s", id)
	var order BroadbandOrderDetail
	if err := authedGet(ctx, path, &order); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(order)
}

func handleBroadbandOrderEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/orders/broadband/%s/events", id)
	var events []OrderEvent
	if err := authedGet(ctx, path, &events); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(events)
}

func handleEthernetOrdersList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flags := listFlagsFromRequest(request)
	orders, err := fetchAll[EthernetOrderListItem](ctx, "/api/orders/ethernet", flags)
	if err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(orders)
}

func handleEthernetOrderGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/orders/ethernet/%s", id)
	var order EthernetOrderDetail
	if err := authedGet(ctx, path, &order); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(order)
}

func handleEthernetOrderEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/orders/ethernet/%s/events", id)
	var events []OrderEvent
	if err := authedGet(ctx, path, &events); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(events)
}

// --- Tickets tools ---

func registerTicketsTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("tickets_list",
			mcp.WithDescription("List support tickets. Supports filtering by status, date range, and search."),
			mcp.WithString("status", mcp.Description("Filter by status (e.g. waiting_for_support)")),
			mcp.WithString("search", mcp.Description("Free-text search")),
			mcp.WithString("from", mcp.Description("From date (YYYY-MM-DD)")),
			mcp.WithString("to", mcp.Description("To date (YYYY-MM-DD)")),
			mcp.WithString("sort", mcp.Description("Field to sort by")),
			mcp.WithString("direction", mcp.Description("Sort direction: ASC or DESC")),
			mcp.WithNumber("limit", mcp.Description("Max results (0 = no limit)")),
		),
		handleTicketsList,
	)

	s.AddTool(
		mcp.NewTool("tickets_get",
			mcp.WithDescription("Get details of a single support ticket."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Ticket ID")),
		),
		handleTicketGet,
	)

	s.AddTool(
		mcp.NewTool("tickets_events",
			mcp.WithDescription("List events/messages for a support ticket."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Ticket ID")),
		),
		handleTicketEvents,
	)

	s.AddTool(
		mcp.NewTool("tickets_create",
			mcp.WithDescription("Create a new support ticket. Returns the ticket ID and portal URL."),
			mcp.WithString("title", mcp.Required(), mcp.Description("Short summary of the issue")),
			mcp.WithString("description", mcp.Required(), mcp.Description("Detailed description of the issue")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Fault type: total_loss, intermittent_connectivity, change_request, query, or network")),
			mcp.WithString("service", mcp.Description("Internet service ID to associate with the ticket")),
			mcp.WithString("cc", mcp.Description("Comma-separated CC email addresses")),
		),
		handleTicketCreate,
	)

	s.AddTool(
		mcp.NewTool("tickets_reply",
			mcp.WithDescription("Send a reply message on an existing support ticket."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Ticket ID")),
			mcp.WithString("message", mcp.Required(), mcp.Description("Message to post")),
		),
		handleTicketReply,
	)
}

func handleTicketsList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flags := listFlagsFromRequest(request)
	tickets, err := fetchAll[TicketListItem](ctx, "/api/tickets", flags)
	if err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(tickets)
}

func handleTicketGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/tickets/%s", id)
	var ticket TicketDetail
	if err := authedGet(ctx, path, &ticket); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(ticket)
}

func handleTicketEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/tickets/%s/events", id)
	var events []TicketEvent
	if err := authedGet(ctx, path, &events); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(events)
}

func handleTicketCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, err := requiredString(request, "title")
	if err != nil {
		return mcpError(err.Error())
	}
	description, err := requiredString(request, "description")
	if err != nil {
		return mcpError(err.Error())
	}
	ticketType, err := requiredString(request, "type")
	if err != nil {
		return mcpError(err.Error())
	}

	req := CreateTicketRequest{
		Title:             title,
		Description:       description,
		Type:              ticketType,
		InternetServiceID: optionalString(request, "service"),
	}

	if cc := optionalString(request, "cc"); cc != "" {
		for _, email := range splitAndTrim(cc) {
			if email != "" {
				req.CarbonCopyEmails = append(req.CarbonCopyEmails, email)
			}
		}
	}

	var resp CreateTicketResponse
	if err := authedPost(ctx, "/api/tickets", req, &resp); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(resp)
}

func handleTicketReply(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcpError(err.Error())
	}
	message, err := requiredString(request, "message")
	if err != nil {
		return mcpError(err.Error())
	}

	path := fmt.Sprintf("/api/tickets/%s/messages", id)
	var resp TicketMessageResponse
	if err := authedPost(ctx, path, TicketMessageRequest{Message: message}, &resp); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(resp)
}

// --- Catalogue tools ---

func registerCatalogueTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("catalogue_additional_products",
			mcp.WithDescription("List additional products available for orders (routers, care levels, IPv4 ranges, onsite install)."),
			mcp.WithString("service_type", mcp.Description("Filter by service type: ethernet or broadband")),
			mcp.WithString("product_type", mcp.Description("Comma-separated product types: onsite_install, ipv4_range, router, care_level")),
		),
		handleCatalogueAdditionalProducts,
	)
}

func handleCatalogueAdditionalProducts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceType := optionalString(request, "service_type")
	productType := optionalString(request, "product_type")

	path := "/api/catalogue/additional-products"
	sep := "?"
	if serviceType != "" {
		path += sep + "service_type=" + url.QueryEscape(serviceType)
		sep = "&"
	}
	if productType != "" {
		path += sep + "product_type=" + url.QueryEscape(productType)
	}

	var products []AdditionalProduct
	if err := authedGet(ctx, path, &products); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(products)
}

// --- Account tools ---

func registerAccountTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("whoami",
			mcp.WithDescription("Show the currently authenticated IP River account."),
		),
		handleWhoami,
	)
}

func handleWhoami(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := newAuthedClient()
	if err != nil {
		return mcpError("Not logged in. Run `ipriver login` in a terminal first.")
	}

	var user struct {
		UUID     string `json:"uuid"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}
	if err := client.Get(ctx, "/api/user", &user); err != nil {
		return mcpError(err.Error())
	}
	return jsonResult(user)
}

// --- Parameter helpers ---

func requiredString(request mcp.CallToolRequest, key string) (string, error) {
	val, ok := request.GetArguments()[key]
	if !ok || val == nil {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	if s == "" {
		return "", fmt.Errorf("parameter %s cannot be empty", key)
	}
	return s, nil
}

func optionalString(request mcp.CallToolRequest, key string) string {
	val, ok := request.GetArguments()[key]
	if !ok || val == nil {
		return ""
	}
	s, _ := val.(string)
	return s
}

func optionalFloat(request mcp.CallToolRequest, key string) (float64, bool) {
	val, ok := request.GetArguments()[key]
	if !ok || val == nil {
		return 0, false
	}
	f, ok := val.(float64)
	return f, ok
}

func listFlagsFromRequest(request mcp.CallToolRequest) ListFlags {
	limit := 0
	if l, ok := optionalFloat(request, "limit"); ok && l > 0 {
		limit = int(l)
	}
	return ListFlags{
		Status:    optionalString(request, "status"),
		Search:    optionalString(request, "search"),
		From:      optionalString(request, "from"),
		To:        optionalString(request, "to"),
		Sort:      optionalString(request, "sort"),
		Direction: optionalString(request, "direction"),
		Limit:     limit,
	}
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
