package main

import (
	"fmt"
	"testing"
)

// TestIsExitCommand checks various exit command cases
func TestIsExitCommand(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"exit", true},
		{"EXIT", true},
		{"quit", true},
		{"QUIT", true},
		{".quit", true},
		{"\\Q", true},
		{":wq", true},
		{"ExIt", true},
		{"Quit", true},
		{"\\q", true},

		// Non-exit commands should return false
		{"hello", false},
		{"shutdown", false},
		{"disconnect", false},
		{"exit now", false}, // "exit" but with extra words
		{"wq:", false},      // Incorrect ":wq" variation
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := isExitCommand(tt.query)
			if result != tt.expected {
				t.Errorf("isExitCommand(%q) = %v; want %v", tt.query, result, tt.expected)
			}
		})
	}
}

// Test getVersionSQL for different database types
func TestGetVersionSQL(t *testing.T) {
	tests := []struct {
		dbType   string
		expected string
	}{
		{"oracle", "SELECT banner FROM v$version"},
		{"sqlserver", "SELECT @@VERSION;"},
		{"snowflake", "SELECT CURRENT_VERSION();"},
		{"sqlite", "SELECT SQLITE_VERSION();"},
		{"postgres", "SELECT CURRENT_SETTING('server_version');"},
		{"mysql", "SELECT VERSION()"},
		{"mariadb", "SELECT VERSION()"}, // MariaDB uses same as MySQL
		{"unknown", "SELECT VERSION()"}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := getVersionSQL(tt.dbType)
			if result != tt.expected {
				t.Errorf("getVersionSQL(%q) = %q; want %q", tt.dbType, result, tt.expected)
			}
		})
	}
}

// Test constructDSN for different database types
func TestConstructDSN(t *testing.T) {
	tests := []struct {
		dbType   string
		user     string
		password string
		host     string
		port     int
		database string
		expected string
	}{
		// PostgreSQL case
		{
			"postgres", "user1", "pass123", "localhost", 5432, "testdb",
			"postgres://user1:pass123@localhost:5432/testdb?sslmode=disable",
		},
		// MySQL/MariaDB case
		{
			"mysql", "user2", "securepass", "db.example.com", 3306, "mydb",
			"user2:securepass@tcp(db.example.com:3306)/mydb",
		},
		// Another MySQL test case
		{
			"mariadb", "admin", "rootpass", "192.168.1.10", 3307, "production",
			"admin:rootpass@tcp(192.168.1.10:3307)/production",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.dbType, tt.database), func(t *testing.T) {
			result := constructDSN(tt.dbType, tt.user, tt.password, tt.host, tt.port, tt.database)
			if result != tt.expected {
				t.Errorf("constructDSN(%q, %q, %q, %q, %d, %q) = %q; want %q",
					tt.dbType, tt.user, tt.password, tt.host, tt.port, tt.database, result, tt.expected)
			}
		})
	}
}

func TestDefaultPortLookup(t *testing.T) {
	expected := map[string]int{
		"mysql":      3306,
		"postgres":   5432,
		"sqlserver":  1433,
		"oracle":     1521,
		"sqlite":     0,
	}

	result := defaultPortLookup()

	// Check if both maps have the same keys and values
	for key, expectedPort := range expected {
		if result[key] != expectedPort {
			t.Errorf("defaultPortLookup()[%q] = %d; want %d", key, result[key], expectedPort)
		}
	}

	// Ensure there are no extra keys in the map
	if len(result) != len(expected) {
		t.Errorf("defaultPortLookup() returned unexpected number of entries: got %d, want %d", len(result), len(expected))
	}
}

// Test getDefaultPort for known and unknown database types
func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		dbType   string
		expected int
	}{
		{"mysql", 3306},
		{"postgres", 5432},
		{"sqlserver", 1433},
		{"oracle", 1521},
		{"sqlite", 0},
		{"unknown", 0}, // Unknown database should return 0
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := getDefaultPort(tt.dbType)
			if result != tt.expected {
				t.Errorf("getDefaultPort(%q) = %d; want %d", tt.dbType, result, tt.expected)
			}
		})
	}
}
