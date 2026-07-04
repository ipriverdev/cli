package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/ipriverdev/cli/internal/api"
	"github.com/ipriverdev/cli/internal/auth"
	"github.com/ipriverdev/cli/internal/config"
	"github.com/spf13/cobra"
)

func newAuthedClient() (*api.Client, error) {
	client, err := auth.NewAuthenticatedClient(config.APIHost())
	if err != nil {
		return nil, fmt.Errorf("not logged in — run `ipriver login` first")
	}

	return client, nil
}

func authedGet(ctx context.Context, path string, out any) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	return client.Get(ctx, path, out)
}

func authedPost(ctx context.Context, path string, body, out any) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	return client.Post(ctx, path, body, out)
}

func authedPostStatus(ctx context.Context, path string, body, out any) (int, error) {
	client, err := newAuthedClient()
	if err != nil {
		return 0, err
	}
	return client.PostStatus(ctx, path, body, out)
}

func authedGetWithStatus(ctx context.Context, path string, out any) (int, error) {
	client, err := newAuthedClient()
	if err != nil {
		return 0, err
	}
	return client.GetWithStatus(ctx, path, out)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type PaginationMeta struct {
	CurrentPage     int  `json:"current_page"`
	HasPreviousPage bool `json:"has_previous_page"`
	HasNextPage     bool `json:"has_next_page"`
	PerPage         int  `json:"per_page"`
	TotalItems      int  `json:"total_items"`
	TotalPages      int  `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

type ListFlags struct {
	Status    string
	Search    string
	From      string
	To        string
	Sort      string
	Direction string
	Limit     int
}

func addListFlags(cmd *cobra.Command) {
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("search", "", "Free-text search")
	cmd.Flags().String("from", "", "Filter results from this date (YYYY-MM-DD)")
	cmd.Flags().String("to", "", "Filter results until this date (YYYY-MM-DD)")
	cmd.Flags().String("sort", "", "Field to sort by")
	cmd.Flags().String("direction", "", "Sort direction (ASC or DESC)")
	cmd.Flags().Int("limit", 0, "Max total results to return (0 = no limit)")
}

func getListFlags(cmd *cobra.Command) ListFlags {
	status, _ := cmd.Flags().GetString("status")
	search, _ := cmd.Flags().GetString("search")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	sort, _ := cmd.Flags().GetString("sort")
	direction, _ := cmd.Flags().GetString("direction")
	limit, _ := cmd.Flags().GetInt("limit")
	return ListFlags{
		Status:    status,
		Search:    search,
		From:      from,
		To:        to,
		Sort:      sort,
		Direction: direction,
		Limit:     limit,
	}
}

func (f ListFlags) toQuery() url.Values {
	q := url.Values{}
	if f.Status != "" {
		q.Set("status", f.Status)
	}
	if f.Search != "" {
		q.Set("search", f.Search)
	}
	if f.From != "" {
		q.Set("from", f.From)
	}
	if f.To != "" {
		q.Set("to", f.To)
	}
	if f.Sort != "" {
		q.Set("sort", f.Sort)
	}
	if f.Direction != "" {
		q.Set("direction", f.Direction)
	}
	return q
}

const (
	maxPageSize = 200
	maxPages    = 500 // safety cap: 500 * 200 = 100k items
)

func fetchAll[T any](ctx context.Context, basePath string, flags ListFlags, extraParams ...func(url.Values)) ([]T, error) {
	client, err := newAuthedClient()
	if err != nil {
		return nil, err
	}

	q := flags.toQuery()
	q.Set("items", strconv.Itoa(maxPageSize))
	for _, fn := range extraParams {
		fn(q)
	}

	var all []T
	page := 1

	for {
		q.Set("page", strconv.Itoa(page))
		path := fmt.Sprintf("%s?%s", basePath, q.Encode())

		var resp PaginatedResponse[T]
		if err := client.Get(ctx, path, &resp); err != nil {
			return nil, err
		}

		all = append(all, resp.Items...)

		if flags.Limit > 0 && len(all) >= flags.Limit {
			all = all[:flags.Limit]
			break
		}

		if !resp.Pagination.HasNextPage {
			break
		}

		if resp.Pagination.TotalPages > 0 && page >= resp.Pagination.TotalPages {
			break
		}

		page++
		if page > maxPages {
			break
		}
	}

	return all, nil
}
