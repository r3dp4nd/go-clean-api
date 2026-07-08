package server

import "testing"

func TestPathParamAfterPrefix(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		prefix     string
		wantValue  string
		wantResult bool
	}{
		{
			name:       "valid product id",
			path:       "/api/v1/products/123",
			prefix:     "/api/v1/products/",
			wantValue:  "123",
			wantResult: true,
		},
		{
			name:       "valid encoded value",
			path:       "/api/v1/products/product%201",
			prefix:     "/api/v1/products/",
			wantValue:  "product 1",
			wantResult: true,
		},
		{
			name:       "empty id",
			path:       "/api/v1/products/",
			prefix:     "/api/v1/products/",
			wantValue:  "",
			wantResult: false,
		},
		{
			name:       "nested path rejected",
			path:       "/api/v1/products/123/details",
			prefix:     "/api/v1/products/",
			wantValue:  "",
			wantResult: false,
		},
		{
			name:       "invalid escaped value",
			path:       "/api/v1/products/%ZZ",
			prefix:     "/api/v1/products/",
			wantValue:  "",
			wantResult: false,
		},
		{
			name:       "wrong prefix",
			path:       "/api/v1/users/123",
			prefix:     "/api/v1/products/",
			wantValue:  "",
			wantResult: false,
		},
		{
			name:       "spaces only rejected",
			path:       "/api/v1/products/%20%20%20",
			prefix:     "/api/v1/products/",
			wantValue:  "",
			wantResult: false,
		},
		{
			name:       "encoded slash rejected",
			path:       "/api/v1/products/abc%2Fdef",
			prefix:     "/api/v1/products/",
			wantValue:  "",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotResult := pathParamAfterPrefix(tt.path, tt.prefix)

			if gotResult != tt.wantResult {
				t.Fatalf("expected result %v, got %v", tt.wantResult, gotResult)
			}

			if gotValue != tt.wantValue {
				t.Fatalf("expected value %q, got %q", tt.wantValue, gotValue)
			}
		})
	}
}
