package waldur

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type paginatedResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

func buildQueryPath(path string, params url.Values) string {
	if len(params) == 0 {
		return path
	}
	return path + "?" + params.Encode()
}

func parseUUIDParam(field, value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid %s: %w", field, err)
	}
	return parsed, nil
}

func classifyTransportError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w: %v", ErrTimeout, err)
	}

	return fmt.Errorf("%w: %v", ErrServerError, err)
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func (c *Client) getPaginated(ctx context.Context, path string, target any) error {
	if err := c.ensureNegotiated(ctx); err != nil {
		return err
	}

	switch out := target.(type) {
	case *[]Category:
		items, err := getPaginatedItems[Category](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	case *[]SupportIssue:
		items, err := getPaginatedItems[SupportIssue](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	case *[]SupportComment:
		items, err := getPaginatedItems[SupportComment](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	case *[]UsageRecord:
		items, err := getPaginatedItems[UsageRecord](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	case *[]offeringResponse:
		items, err := getPaginatedItems[offeringResponse](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	case *[]Order:
		items, err := getPaginatedItems[Order](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	case *[]Resource:
		items, err := getPaginatedItems[Resource](ctx, c, path)
		if err != nil {
			return err
		}
		*out = items
		return nil
	default:
		return fmt.Errorf("unsupported paginated target %T", target)
	}
}

func getPaginatedItems[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	items := make([]T, 0)
	nextURL := path
	seen := map[string]struct{}{}

	for pageCount := 0; pageCount < 1000; pageCount++ {
		respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, err
		}
		if statusCode != http.StatusOK {
			return nil, mapHTTPError(statusCode, respBody)
		}

		var direct []T
		if err := json.Unmarshal(respBody, &direct); err == nil {
			items = append(items, direct...)
			return items, nil
		}

		var wrapped paginatedResponse[T]
		if err := json.Unmarshal(respBody, &wrapped); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		}

		items = append(items, wrapped.Results...)
		if wrapped.Next == "" {
			return items, nil
		}
		if _, exists := seen[wrapped.Next]; exists {
			return nil, fmt.Errorf("%w: pagination loop detected at %s", ErrInvalidResponse, wrapped.Next)
		}
		seen[wrapped.Next] = struct{}{}
		nextURL = wrapped.Next
	}

	return nil, fmt.Errorf("%w: pagination exceeded 1000 pages", ErrInvalidResponse)
}
