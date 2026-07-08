package product

import "testing"

func TestPostgresSortExpression(t *testing.T) {
	tests := []struct {
		name       string
		sortField  string
		wantResult string
	}{
		{
			name:       "sort by id",
			sortField:  SortFieldID,
			wantResult: "id",
		},
		{
			name:       "sort by name",
			sortField:  SortFieldName,
			wantResult: "lower(name)",
		},
		{
			name:       "sort by price",
			sortField:  SortFieldPrice,
			wantResult: "price",
		},
		{
			name:       "sort by created at",
			sortField:  SortFieldCreatedAt,
			wantResult: "created_at",
		},
		{
			name:       "sort by updated at",
			sortField:  SortFieldUpdatedAt,
			wantResult: "updated_at",
		},
		{
			name:       "unknown sort falls back to id",
			sortField:  "unknown",
			wantResult: "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := postgresSortExpression(tt.sortField)

			if gotResult != tt.wantResult {
				t.Fatalf("expected %q, got %q", tt.wantResult, gotResult)
			}
		})
	}
}

func TestPostgresSortDirection(t *testing.T) {
	tests := []struct {
		name       string
		order      string
		wantResult string
	}{
		{
			name:       "ascending",
			order:      SortOrderAsc,
			wantResult: "ASC",
		},
		{
			name:       "descending",
			order:      SortOrderDesc,
			wantResult: "DESC",
		},
		{
			name:       "unknown order falls back to asc",
			order:      "random",
			wantResult: "ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := postgresSortDirection(tt.order)

			if gotResult != tt.wantResult {
				t.Fatalf("expected %q, got %q", tt.wantResult, gotResult)
			}
		})
	}
}
