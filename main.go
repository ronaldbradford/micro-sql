package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var rowLimit int
var executionCount int
var dbType string

func main() {
	// Detect database type based on the program name
	progName := strings.ToLower(os.Args[0])
	switch {
	case strings.Contains(progName, "psql"):
		dbType = "postgresql"
	default:
		dbType = "mysql" // Default to MySQL
	}

	// Define command-line flags
	user := flag.String("u", "", "Database username")
	password := flag.String("p", "", "Database password")
	host := flag.String("h", "127.0.0.1", "Database host")
	port := flag.Int("P", defaultPort(), "Database port")
	flag.IntVar(&rowLimit, "l", 10, "Number of rows to display before truncation (default: 10)")
	flag.IntVar(&executionCount, "c", 1, "Number of times to execute the query (default: 1)")
	flag.Parse()

	// Ensure a database name is provided
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Database name is required.")
		os.Exit(1)
	}
	database := args[0]

	// Validate required parameters
	if *user == "" {
		fmt.Println("Error: Database username (-u) is required.")
		os.Exit(1)
	}

	// Construct DSN
	dsn := constructDSN(dbType, *user, *password, *host, *port, database)

	// Open database connection
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	fmt.Printf("Connected to %s database '%s'!\n", dbType, database)

	// Handle SIGINT (^C) signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived SIGINT, exiting...")
		db.Close()
		os.Exit(0)
	}()

	reader := bufio.NewReader(os.Stdin)

	// Query input loop
	for {
		fmt.Printf("micro-%s (%s)> ", dbType, time.Now().Format("15:04:05"))

		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		query = strings.TrimSpace(query)

		if isExitCommand(query) {
			fmt.Println("Goodbye!")
			break
		}

		if handleMicroCommand(query) {
			continue
		}

		if strings.EqualFold(query, "HELP") {
			displayHelp()
			continue
		}

		if strings.HasPrefix(strings.ToLower(query), "select") {
			executeQuery(db, query, rowLimit, executionCount)
		} else {
			fmt.Println("Only SELECT statements are allowed. Type 'HELP' for available commands.")
		}
	}
}

// Default port based on database type
func defaultPort() int {
	if strings.Contains(os.Args[0], "psql") {
		return 5432
	}
	return 3306
}

// Construct DSN for MySQL and PostgreSQL
func constructDSN(dbType, user, password, host string, port int, database string) string {
	if dbType == "mysql" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database)
}

// Check exit command (case-insensitive, supports multiple ways to exit)
func isExitCommand(input string) bool {
	lowerInput := strings.ToLower(strings.TrimSpace(input))
	exitCommands := []string{"exit", "quit", "\\q", ":wq"}

	for _, cmd := range exitCommands {
		if lowerInput == cmd {
			return true
		}
	}
	return false
}

// Handle SET MICRO commands
func handleMicroCommand(input string) bool {
	re := regexp.MustCompile(`\s*=\s*`)
	input = re.ReplaceAllString(input, " ")
	words := strings.Fields(strings.ToUpper(input))

	if len(words) != 4 || words[0] != "SET" || words[1] != "MICRO" {
		return false
	}

	value, err := parseInt(words[3])
	if err != nil || value < 1 {
		fmt.Println("Invalid value. Must be a positive integer.")
		return true
	}

	switch words[2] {
	case "COUNT":
		executionCount = value
		fmt.Printf("Execution count set to %d\n", executionCount)
	case "LIMIT":
		rowLimit = value
		fmt.Printf("Row limit set to %d\n", rowLimit)
	default:
		return false
	}

	return true
}

// Convert string to integer
func parseInt(s string) (int, error) {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	return value, err
}

// Display help
func displayHelp() {
	fmt.Println("\nAvailable Commands:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("HELP                 - Display this help message\n")
	fmt.Printf("EXIT                 - Exit the application\n")
	fmt.Printf("SET MICRO COUNT=N    - Set the execution count for queries (Currently %d)\n", executionCount)
	fmt.Printf("SET MICRO LIMIT=N    - Set the maximum number of rows displayed (Currently %d)\n", rowLimit)
	fmt.Printf("SELECT ...           - Execute a SELECT query\n")
	fmt.Println(strings.Repeat("-", 50))
}

// Execute SELECT queries with separate query and response time
func executeQuery(db *sql.DB, query string, rowLimit, count int) {
	for i := 0; i < count; i++ {
		totalStart := time.Now()

		queryStart := time.Now()
		rows, err := db.Query(query)
		queryElapsed := time.Since(queryStart)

		if err != nil {
			fmt.Printf("Query Error: %v\n", err)
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			fmt.Printf("Error fetching columns: %v\n", err)
			return
		}

		rowCount := 0
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println(strings.Join(columns, "\t"))
		fmt.Println(strings.Repeat("-", 50))

		for rows.Next() {
			rowCount++
			if rowCount <= rowLimit {
				fmt.Println("[Row data]")
			}
		}

		if rowCount > rowLimit {
			fmt.Printf("[...] Output truncated at %d rows.\n", rowLimit)
		}

		totalElapsed := time.Since(totalStart)

		fmt.Printf("%d rows (%.6f ms query, %.6f ms result)\n",
			rowCount,
			float64(queryElapsed.Nanoseconds())/1e6,
			float64(totalElapsed.Nanoseconds())/1e6,
		)
	}
}
