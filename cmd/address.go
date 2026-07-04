package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ipriverdev/cli/internal/app"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

type FlexString string

func (f *FlexString) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*f = ""
	} else {
		*f = FlexString(s)
	}
	return nil
}

type Address struct {
	UPRN           FlexString `json:"uprn"`
	ALK            string     `json:"alk"`
	District       string     `json:"district"`
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
	Postcode       string     `json:"postcode"`
	Address        string     `json:"address"`
	AddressLine1   string     `json:"address_line_one"`
	AddressLine2   string     `json:"address_line_two"`
	AddressLine3   string     `json:"address_line_three"`
	AddressLine4   string     `json:"address_line_four"`
	Town           string     `json:"town"`
	Street         string     `json:"street"`
	County         string     `json:"county"`
	Organisation   string     `json:"organisation"`
	BuildingName   string     `json:"building_name"`
	BuildingNumber string     `json:"building_number"`
}

func addressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "Look up addresses",
	}

	cmd.AddCommand(addressPostcodeCmd())
	cmd.AddCommand(addressUPRNCmd())

	return cmd
}

func addressPostcodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "postcode <postcode>",
		Short: "Search addresses by postcode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddressPostcode(cmd.Context(), args[0])
		},
	}
}

func addressUPRNCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uprn <uprn>",
		Short: "Get a single address by UPRN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddressUPRN(cmd.Context(), args[0])
		},
	}
}

func runAddressPostcode(ctx context.Context, postcode string) error {
	var addresses []Address
	path := fmt.Sprintf("/api/address/postcode/%s", url.PathEscape(postcode))
	sp := ui.NewSpinner("Searching addresses…")
	err := authedGet(ctx, path, &addresses)
	sp.Stop()
	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		ui.Info("No addresses found.")
		return nil
	}

	if app.JSON() {
		return printJSON(addresses)
	}

	printAddressTable(addresses)
	return nil
}

func runAddressUPRN(ctx context.Context, uprn string) error {
	var addr Address
	path := fmt.Sprintf("/api/address/uprn/%s", url.PathEscape(uprn))
	sp := ui.NewSpinner("Fetching address…")
	err := authedGet(ctx, path, &addr)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(addr)
	}

	printAddressDetail(addr)
	return nil
}

func printAddressTable(addresses []Address) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  #\tADDRESS\tUPRN\tALK\tDC\n")
	for i, a := range addresses {
		fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n", i+1, a.Address, string(a.UPRN), a.ALK, a.District)
	}
	_ = w.Flush()
	fmt.Printf("\n%d address(es) found.\n", len(addresses))
}

func printAddressDetail(addr Address) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if addr.UPRN != "" {
		fmt.Fprintf(w, "UPRN:\t%s\n", string(addr.UPRN))
	}
	fmt.Fprintf(w, "ALK:\t%s\n", addr.ALK)
	fmt.Fprintf(w, "Address:\t%s\n", addr.Address)
	if addr.BuildingName != "" {
		fmt.Fprintf(w, "Building:\t%s\n", addr.BuildingName)
	}
	if addr.BuildingNumber != "" {
		fmt.Fprintf(w, "Number:\t%s\n", addr.BuildingNumber)
	}
	fmt.Fprintf(w, "Street:\t%s\n", addr.Street)
	fmt.Fprintf(w, "Town:\t%s\n", addr.Town)
	if addr.County != "" {
		fmt.Fprintf(w, "County:\t%s\n", addr.County)
	}
	fmt.Fprintf(w, "Postcode:\t%s\n", addr.Postcode)
	if addr.Organisation != "" {
		fmt.Fprintf(w, "Organisation:\t%s\n", addr.Organisation)
	}
	if addr.District != "" {
		fmt.Fprintf(w, "District:\t%s\n", addr.District)
	}
	_ = w.Flush()
}
