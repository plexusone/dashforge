package dashboardir

import "testing"

func TestAnalyticsSourceConfigCredentialReference(t *testing.T) {
	tests := []struct {
		name string
		cfg  AnalyticsSourceConfig
		want string
	}{
		{
			name: "credential ref wins",
			cfg: AnalyticsSourceConfig{ //nolint:gosec // G101: CredentialRef is a reference locator (sql://...), not a hardcoded credential
				CredentialRef: "sql://analytics-sources/source",
				DSNRef:        "env://SOURCE_DSN",
			},
			want: "sql://analytics-sources/source",
		},
		{
			name: "dsn ref fallback",
			cfg:  AnalyticsSourceConfig{DSNRef: "env://SOURCE_DSN"},
			want: "env://SOURCE_DSN",
		},
		{
			name: "empty",
			cfg:  AnalyticsSourceConfig{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.CredentialReference(); got != tt.want {
				t.Fatalf("CredentialReference() = %q, want %q", got, tt.want)
			}
		})
	}
}
