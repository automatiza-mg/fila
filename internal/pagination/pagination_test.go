package pagination

import (
	"net/http"
	"testing"
)

func TestOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{"first page", 1, 20, 0},
		{"second page", 2, 20, 20},
		{"tenth page", 10, 50, 450},
		{"large values", 100, 100, 9900},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Offset(tt.page, tt.limit)
			if got != tt.expected {
				t.Errorf("Offset(%d, %d) = %d, want %d", tt.page, tt.limit, got, tt.expected)
			}
		})
	}
}

func TestNewResult(t *testing.T) {
	data := []string{"a", "b", "c"}

	tests := []struct {
		name            string
		totalCount      int
		limit           int
		page            int
		expectedPages   int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{"first page with more", 100, 20, 1, 5, true, false},
		{"middle page", 100, 20, 3, 5, true, true},
		{"last page", 100, 20, 5, 5, false, true},
		{"single page", 15, 20, 1, 1, false, false},
		{"exact fit", 60, 20, 2, 3, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResult(data, tt.page, tt.totalCount, tt.limit)

			if result.TotalPages != tt.expectedPages {
				t.Errorf("TotalPages = %d, want %d", result.TotalPages, tt.expectedPages)
			}
			if result.HasNext != tt.expectedHasNext {
				t.Errorf("HasNext = %v, want %v", result.HasNext, tt.expectedHasNext)
			}
			if result.HasPrev != tt.expectedHasPrev {
				t.Errorf("HasPrev = %v, want %v", result.HasPrev, tt.expectedHasPrev)
			}
			if result.CurrentPage != tt.page {
				t.Errorf("CurrentPage = %d, want %d", result.CurrentPage, tt.page)
			}
			if result.Limit != tt.limit {
				t.Errorf("Limit = %d, want %d", result.Limit, tt.limit)
			}
			if result.TotalCount != tt.totalCount {
				t.Errorf("TotalCount = %d, want %d", result.TotalCount, tt.totalCount)
			}
		})
	}
}

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedPage  int
		expectedLimit int
	}{
		// Valid cases
		{"valid page and limit", "?page=2&limit=50", 2, 50},
		{"valid page only", "?page=5", 5, DefaultLimit},
		{"valid limit only", "?limit=30", DefaultPage, 30},
		{"empty query", "", DefaultPage, DefaultLimit},

		// Invalid page values (should use default page)
		{"invalid page - non-numeric", "?page=abc&limit=20", DefaultPage, 20},
		{"invalid page - float", "?page=1.5&limit=20", DefaultPage, 20},
		{"negative page", "?page=-5", MinPage, DefaultLimit},
		{"zero page", "?page=0", MinPage, DefaultLimit},

		// Invalid limit values (should use default limit)
		{"invalid limit - non-numeric", "?page=2&limit=xyz", 2, DefaultLimit},
		{"invalid limit - float", "?page=2&limit=25.5", 2, DefaultLimit},
		{"negative limit", "?page=2&limit=-10", 2, MinLimit},
		{"zero limit", "?page=2&limit=0", 2, MinLimit},

		// Clamping tests
		{"limit exceeds max", "?page=1&limit=200", 1, MaxLimit},
		{"limit exceeds max by little", "?page=1&limit=51", 1, MaxLimit},
		{"limit at max", "?page=1&limit=50", 1, MaxLimit},
		{"limit at min", "?page=1&limit=1", 1, 1},

		// Multiple invalid parameters
		{"all invalid", "?page=abc&limit=xyz", DefaultPage, DefaultLimit},
		{"page invalid, limit exceeds max", "?page=xyz&limit=500", DefaultPage, MaxLimit},

		// Large valid values
		{"large page", "?page=9999", 9999, DefaultLimit},
		{"large limit (clamped)", "?page=1&limit=999", 1, MaxLimit},

		// Edge cases
		{"page=1 explicitly", "?page=1", 1, DefaultLimit},
		{"limit=20 explicitly", "?limit=20", DefaultPage, 20},
		{"whitespace in value", "?page= 2", DefaultPage, DefaultLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com/api"+tt.query, nil)
			params := ParseQuery(req)

			if params.Page != tt.expectedPage {
				t.Errorf("Page = %d, want %d (query: %s)", params.Page, tt.expectedPage, tt.query)
			}
			if params.Limit != tt.expectedLimit {
				t.Errorf("Limit = %d, want %d (query: %s)", params.Limit, tt.expectedLimit, tt.query)
			}
		})
	}
}

func TestParseQuery_Boundaries(t *testing.T) {
	// Test that min/max clamping works correctly with builtin functions
	tests := []struct {
		name          string
		query         string
		expectedPage  int
		expectedLimit int
	}{
		{"page below min", "?page=-100&limit=50", MinPage, 50},
		{"limit below min", "?page=2&limit=-50", 2, MinLimit},
		{"limit above max", "?page=2&limit=1000", 2, MaxLimit},
		{"all at boundaries", "?page=1&limit=50", 1, MaxLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com/api"+tt.query, nil)
			params := ParseQuery(req)

			if params.Page != tt.expectedPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.expectedPage)
			}
			if params.Limit != tt.expectedLimit {
				t.Errorf("Limit = %d, want %d", params.Limit, tt.expectedLimit)
			}
		})
	}
}
