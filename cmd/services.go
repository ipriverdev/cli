package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/ipriverdev/cli/internal/app"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

type InternetService struct {
	ID             string                 `json:"id"`
	Reference      string                 `json:"reference"`
	NetworkType    string                 `json:"network_type"`
	Technology     string                 `json:"technology"`
	Contract       *ServiceContract       `json:"contract,omitempty"`
	EndUserDetails *ServiceEndUserDetails `json:"end_user_details,omitempty"`
}

type ServiceContract struct {
	TermInMonths int    `json:"term_in_months"`
	StartDate    string `json:"start_date,omitempty"`
}

type ServiceEndUserDetails struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Postcode  string  `json:"postcode"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func servicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage internet services",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all internet services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServicesList(cmd)
		},
	}
	addListFlags(listCmd)
	listCmd.Flags().String("technology", "", "Filter by technology (ethernet, broadband, p2p)")
	listCmd.Flags().String("network-type", "", "Filter by network type (on_net, off_net)")

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single internet service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServicesGet(cmd.Context(), args[0])
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(getCmd)
	return cmd
}

func runServicesList(cmd *cobra.Command) error {
	flags := getListFlags(cmd)
	technology, _ := cmd.Flags().GetString("technology")
	networkType, _ := cmd.Flags().GetString("network-type")

	sp := ui.NewSpinner("Fetching services…")
	services, err := fetchAll[InternetService](cmd.Context(), "/api/internet-services", flags, func(q url.Values) {
		if technology != "" {
			q.Add("technology[]", technology)
		}
		if networkType != "" {
			q.Add("networkType[]", networkType)
		}
	})
	sp.Stop()

	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(services)
	}

	if len(services) == 0 {
		ui.Info("No services found.")
		return nil
	}

	printServicesTable(services)
	return nil
}

func runServicesGet(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/internet-services/%s", id)
	var svc InternetService
	sp := ui.NewSpinner("Fetching service…")
	err := authedGet(ctx, path, &svc)
	sp.Stop()
	if err != nil {
		return err
	}

	if app.JSON() {
		return printJSON(svc)
	}

	printServiceDetail(svc)
	return nil
}

func printServicesTable(services []InternetService) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  REFERENCE\tTECHNOLOGY\tNETWORK\tEND USER\tPOSTCODE\n")
	for _, s := range services {
		endUser := ""
		postcode := ""
		if s.EndUserDetails != nil {
			endUser = s.EndUserDetails.Name
			postcode = s.EndUserDetails.Postcode
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n", s.Reference, s.Technology, s.NetworkType, endUser, postcode)
	}
	w.Flush()
	fmt.Printf("\n%d service(s) found.\n", len(services))
}

func printServiceDetail(svc InternetService) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%s\n", svc.ID)
	fmt.Fprintf(w, "Reference:\t%s\n", svc.Reference)
	fmt.Fprintf(w, "Technology:\t%s\n", svc.Technology)
	fmt.Fprintf(w, "Network Type:\t%s\n", svc.NetworkType)
	if svc.Contract != nil {
		fmt.Fprintf(w, "Term:\t%d months\n", svc.Contract.TermInMonths)
		if svc.Contract.StartDate != "" {
			fmt.Fprintf(w, "Start Date:\t%s\n", svc.Contract.StartDate)
		}
	}
	if svc.EndUserDetails != nil {
		fmt.Fprintf(w, "End User:\t%s\n", svc.EndUserDetails.Name)
		fmt.Fprintf(w, "Address:\t%s\n", svc.EndUserDetails.Address)
		fmt.Fprintf(w, "Postcode:\t%s\n", svc.EndUserDetails.Postcode)
	}
	w.Flush()
}
