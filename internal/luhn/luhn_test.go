package luhn

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid number 79927398713",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "valid number 12345678903",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "valid number 4539578763621486",
			number: "4539578763621486",
			want:   true,
		},
		{
			name:   "invalid number 12345678904",
			number: "12345678904",
			want:   false,
		},
		{
			name:   "invalid number 1234567890",
			number: "1234567890",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "non-digit characters",
			number: "1234abc5678",
			want:   false,
		},
		{
			name:   "single digit 0",
			number: "0",
			want:   true,
		},
		{
			name:   "single digit 1",
			number: "1",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.number)
			if got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestIsValidOrderNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid order number",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "valid order number 12345678903",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "invalid order number",
			number: "12345678904",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "with letters",
			number: "123abc",
			want:   false,
		},
		{
			name:   "with spaces",
			number: "123 456",
			want:   false,
		},
		{
			name:   "with special characters",
			number: "123-456",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidOrderNumber(tt.number)
			if got != tt.want {
				t.Errorf("IsValidOrderNumber(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}
