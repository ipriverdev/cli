package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ipriverdev/cli/internal/app"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

type AvailabilityCheckRequest struct {
	Postcode       string   `json:"postcode,omitempty"`
	UPRN           string   `json:"uprn,omitempty"`
	ALK            string   `json:"alk,omitempty"`
	DC             string   `json:"dc,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	Address        string   `json:"address,omitempty"`
	BuildingNumber string   `json:"building_number,omitempty"`
	BuildingName   string   `json:"building_name,omitempty"`
	Street         string   `json:"street,omitempty"`
	Town           string   `json:"town,omitempty"`
	County         string   `json:"county,omitempty"`
}

type AvailabilityCheckResponse struct {
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
}

type AvailabilityCheckStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type AvailabilityCheckDetail struct {
	ID               string                    `json:"id"`
	Reference        string                    `json:"reference"`
	CreatedAt        string                    `json:"created_at"`
	Status           string                    `json:"status"`
	User             *AvailabilityCheckUser    `json:"user,omitempty"`
	SearchParameters *AvailabilitySearchParams `json:"search_parameters,omitempty"`
}

type AvailabilityCheckUser struct {
	FullName string                    `json:"full_name"`
	Email    string                    `json:"email"`
	Company  *AvailabilityCheckCompany `json:"company,omitempty"`
}

type AvailabilityCheckCompany struct {
	FullName string `json:"full_name"`
}

type AvailabilitySearchParams struct {
	Postcode  string   `json:"postcode"`
	ALK       string   `json:"alk,omitempty"`
	DC        string   `json:"dc,omitempty"`
	UPRN      string   `json:"uprn,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Address   string   `json:"address,omitempty"`
}

type ProductPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Money    string `json:"money"`
	Decimal  string `json:"decimal"`
}

type SalePrice struct {
	InstallPrice ProductPrice `json:"install_price"`
	AnnualPrice  ProductPrice `json:"annual_price"`
	MonthlyPrice ProductPrice `json:"monthly_price"`
	Currency     string       `json:"currency"`
}

type AvailableProduct struct {
	ID                   string    `json:"id"`
	Carrier              string    `json:"carrier"`
	Technology           string    `json:"technology"`
	ProductName          string    `json:"product_name"`
	TermLengthInMonths   *int      `json:"term_length_in_months"`
	Bearer               *int      `json:"bearer"`
	Bandwidth            *int      `json:"bandwidth"`
	DownloadSpeed        *float64  `json:"download_speed"`
	UploadSpeed          *float64  `json:"upload_speed"`
	MinimumDownloadSpeed *float64  `json:"minimum_download_speed"`
	MaximumDownloadSpeed *float64  `json:"maximum_download_speed"`
	MinimumUploadSpeed   *float64  `json:"minimum_upload_speed"`
	MaximumUploadSpeed   *float64  `json:"maximum_upload_speed"`
	SalePrice            SalePrice `json:"sale_price"`
}

type CheckResult struct {
	CheckID  string             `json:"check_id"`
	Status   string             `json:"status"`
	Products []AvailableProduct `json:"products"`
}

func checkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check product availability for an address",
		Long: `Run an availability check for an address. By default, submits the check,
polls until complete, and returns the list of available products.

Example:
  ipriver check --postcode "SW1A 1AA" --uprn 10033544614 --format=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd)
		},
	}

	cmd.Flags().String("postcode", "", "Postcode to check")
	cmd.Flags().String("uprn", "", "Unique Property Reference Number")
	cmd.Flags().String("alk", "", "Address Location Key (pair with --dc)")
	cmd.Flags().String("dc", "", "District Code (pair with --alk)")
	cmd.Flags().Float64("lat", 0, "Latitude")
	cmd.Flags().Float64("lon", 0, "Longitude")
	cmd.Flags().String("address", "", "Full address string")
	cmd.Flags().Bool("no-wait", false, "Submit only, return check ID without waiting")
	cmd.Flags().Int("timeout", 120, "Max seconds to wait for completion")

	cmd.AddCommand(checkStatusCmd())
	cmd.AddCommand(checkGetCmd())
	cmd.AddCommand(checkProductsCmd())

	return cmd
}

func checkStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <id>",
		Short: "Get status of an availability check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckStatus(cmd.Context(), args[0])
		},
	}
}

func checkGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get details of an availability check request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckGet(cmd.Context(), args[0])
		},
	}
}

func checkProductsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "products <id>",
		Short: "Get products for a completed availability check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckProducts(cmd.Context(), args[0])
		},
	}
}

func runCheck(cmd *cobra.Command) error {
	ctx := cmd.Context()

	postcode, _ := cmd.Flags().GetString("postcode")
	uprn, _ := cmd.Flags().GetString("uprn")
	alk, _ := cmd.Flags().GetString("alk")
	dc, _ := cmd.Flags().GetString("dc")
	lat, _ := cmd.Flags().GetFloat64("lat")
	lon, _ := cmd.Flags().GetFloat64("lon")
	address, _ := cmd.Flags().GetString("address")
	noWait, _ := cmd.Flags().GetBool("no-wait")
	timeout, _ := cmd.Flags().GetInt("timeout")

	if postcode == "" && uprn == "" {
		return fmt.Errorf("at least --postcode or --uprn is required")
	}

	req := AvailabilityCheckRequest{
		Postcode: postcode,
		UPRN:     uprn,
		ALK:      alk,
		DC:       dc,
		Address:  address,
	}
	if cmd.Flags().Changed("lat") || cmd.Flags().Changed("lon") {
		req.Latitude = &lat
		req.Longitude = &lon
	}

	sp := ui.NewSpinner("Submitting availability check…")
	var resp AvailabilityCheckResponse
	status, err := authedPostStatus(ctx, "/api/availability-checker", req, &resp)
	sp.Stop()
	if err != nil {
		return err
	}

	if status != http.StatusAccepted {
		return fmt.Errorf("unexpected status %d", status)
	}

	if noWait {
		return printJSON(map[string]string{
			"check_id": resp.ID,
			"status":   "PENDING",
		})
	}

	products, finalStatus, err := pollAndFetchProducts(ctx, resp.ID, timeout)
	if err != nil {
		return err
	}

	result := CheckResult{
		CheckID:  resp.ID,
		Status:   finalStatus,
		Products: products,
	}

	if app.JSON() {
		return printJSON(result)
	}

	printCheckResult(result)
	return nil
}

func pollAndFetchProducts(ctx context.Context, checkID string, timeoutSecs int) ([]AvailableProduct, string, error) {
	deadline := time.Now().Add(time.Duration(timeoutSecs) * time.Second)
	statusPath := fmt.Sprintf("/api/availability-checker/%s/status", checkID)
	retryInterval := 5 * time.Second

	sp := ui.NewSpinner("Waiting for availability check to complete…")
	defer sp.Stop()

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
			sp.Stop()
			return fetchProducts(ctx, checkID)
		case httpStatus == http.StatusGatewayTimeout || statusResp.Status == "TIMED_OUT":
			sp.Stop()
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

func fetchProducts(ctx context.Context, checkID string) ([]AvailableProduct, string, error) {
	path := fmt.Sprintf("/api/availability-checker/%s/products", checkID)
	var products []AvailableProduct
	if err := authedGet(ctx, path, &products); err != nil {
		return nil, "COMPLETED", err
	}
	return products, "COMPLETED", nil
}

func runCheckStatus(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/availability-checker/%s/status", id)
	var resp AvailabilityCheckStatus
	sp := ui.NewSpinner("Checking status…")
	_, err := authedGetWithStatus(ctx, path, &resp)
	sp.Stop()
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func runCheckGet(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/availability-checker/%s", id)
	var resp AvailabilityCheckDetail
	sp := ui.NewSpinner("Fetching check details…")
	err := authedGet(ctx, path, &resp)
	sp.Stop()
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func runCheckProducts(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/availability-checker/%s/products", id)
	var products []AvailableProduct
	sp := ui.NewSpinner("Fetching products…")
	err := authedGet(ctx, path, &products)
	sp.Stop()
	if err != nil {
		return err
	}
	return printJSON(products)
}

func printCheckResult(result CheckResult) {
	if result.Status == "TIMED_OUT" {
		ui.Warn("Check timed out. Partial results may be shown.")
	}

	if len(result.Products) == 0 {
		ui.Info("No products available for this address.")
		return
	}

	fmt.Printf("Check ID: %s\n", result.CheckID)
	fmt.Printf("Status:   %s\n", result.Status)
	fmt.Printf("Products: %d found\n\n", len(result.Products))

	for i, p := range result.Products {
		fmt.Printf("  %d. %s\n", i+1, p.ProductName)
		fmt.Printf("     Carrier:    %s\n", p.Carrier)
		fmt.Printf("     Technology: %s\n", p.Technology)
		if p.Bearer != nil {
			fmt.Printf("     Bearer:     %d Mbps\n", *p.Bearer)
		}
		if p.Bandwidth != nil {
			fmt.Printf("     Bandwidth:  %d Mbps\n", *p.Bandwidth)
		}
		if p.DownloadSpeed != nil {
			fmt.Printf("     Download:   %.0f Mbps\n", *p.DownloadSpeed)
		}
		if p.UploadSpeed != nil {
			fmt.Printf("     Upload:     %.0f Mbps\n", *p.UploadSpeed)
		}
		if p.TermLengthInMonths != nil {
			fmt.Printf("     Term:       %d months\n", *p.TermLengthInMonths)
		}
		fmt.Printf("     Monthly:    %s\n", p.SalePrice.MonthlyPrice.Money)
		fmt.Printf("     Install:    %s\n", p.SalePrice.InstallPrice.Money)
		fmt.Println()
	}
}
