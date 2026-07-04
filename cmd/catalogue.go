package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ipriverdev/cli/internal/app"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

type AdditionalProduct struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Type             string       `json:"type"`
	Technology       string       `json:"technology,omitempty"`
	Carrier          string       `json:"carrier,omitempty"`
	InstallSalePrice ProductPrice `json:"install_sale_price"`
	AnnualSalePrice  ProductPrice `json:"annual_sale_price"`
	MonthlySalePrice ProductPrice `json:"monthly_sale_price"`
}

func catalogueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalogue",
		Short: "Browse the product catalogue",
	}

	apCmd := &cobra.Command{
		Use:   "additional-products",
		Short: "List additional products (routers, care levels, IPv4 ranges, onsite install)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdditionalProducts(cmd)
		},
	}
	apCmd.Flags().String("service-type", "", "Filter by service type (ethernet, broadband)")
	apCmd.Flags().String("product-type", "", "Comma-separated product types (onsite_install, ipv4_range, router, care_level)")

	cmd.AddCommand(apCmd)
	return cmd
}

func runAdditionalProducts(cmd *cobra.Command) error {
	serviceType, _ := cmd.Flags().GetString("service-type")
	productType, _ := cmd.Flags().GetString("product-type")

	path := "/api/catalogue/additional-products"
	sep := "?"
	if serviceType != "" {
		path += sep + "service_type=" + serviceType
		sep = "&"
	}
	if productType != "" {
		path += sep + "product_type=" + productType
	}

	var products []AdditionalProduct
	sp := ui.NewSpinner("Fetching additional products…")
	err := authedGet(cmd.Context(), path, &products)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(products)
	}

	if len(products) == 0 {
		ui.Info("No additional products found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  NAME\tTYPE\tMONTHLY\tINSTALL\n")
	for _, p := range products {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", p.Name, p.Type, p.MonthlySalePrice.Money, p.InstallSalePrice.Money)
	}
	_ = w.Flush()
	fmt.Printf("\n%d product(s) found.\n", len(products))
	return nil
}
