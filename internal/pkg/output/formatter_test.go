package output

import (
	"testing"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		color      string
		noHeaders  bool
		wantColor  bool
		wantFormat string
	}{
		{
			name:       "table auto color",
			format:     "table",
			color:      "auto",
			noHeaders:  false,
			wantFormat: "table",
		},
		{
			name:       "json no color",
			format:     "json",
			color:      "never",
			noHeaders:  false,
			wantColor:  false,
			wantFormat: "json",
		},
		{
			name:       "table always color",
			format:     "table",
			color:      "always",
			noHeaders:  false,
			wantColor:  true,
			wantFormat: "table",
		},
		{
			name:       "table no headers",
			format:     "table",
			color:      "auto",
			noHeaders:  true,
			wantFormat: "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.format, tt.color, tt.noHeaders)
			if f.Format != tt.wantFormat {
				t.Errorf("NewFormatter() Format = %v, want %v", f.Format, tt.wantFormat)
			}
			if tt.color == "never" && f.Color != false {
				t.Errorf("NewFormatter() Color = %v, want false", f.Color)
			}
			if tt.color == "always" && f.Color != true {
				t.Errorf("NewFormatter() Color = %v, want true", f.Color)
			}
			if f.NoHeaders != tt.noHeaders {
				t.Errorf("NewFormatter() NoHeaders = %v, want %v", f.NoHeaders, tt.noHeaders)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"json valid", "json", false},
		{"table valid", "table", false},
		{"yaml invalid", "yaml", true},
		{"csv invalid", "csv", true},
		{"empty invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
