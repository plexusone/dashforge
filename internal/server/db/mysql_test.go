package db

import "testing"

func TestMySQLURLToDSN(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "url with user and port",
			in:   "mysql://root@127.0.0.1:13308/uiforge",
			want: "root@tcp(127.0.0.1:13308)/uiforge",
		},
		{
			name: "url with password and query",
			in:   "mysql://user:pass@db.example.com/uiforge?tls=true", // #nosec G101 -- placeholder credentials in a parser test
			want: "user:pass@tcp(db.example.com:3306)/uiforge?tls=true",
		},
		{
			name: "raw driver dsn passes through",
			in:   "root:@tcp(127.0.0.1:13308)/uiforge",
			want: "root:@tcp(127.0.0.1:13308)/uiforge",
		},
		{
			name:    "missing database name",
			in:      "mysql://root@127.0.0.1:13308/",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mysqlURLToDSN(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("mysqlURLToDSN(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
