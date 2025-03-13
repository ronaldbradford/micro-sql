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
)

var rowLimit int
var executionCount int

func main() {
	// Define command-line flags
	user := flag.String("u", "", "MySQL username")
	password := flag.String("p", "", "MySQL password")
	host := flag.String("h", "127.0.0.1", "MySQL host")
	port := flag.Int("P", 3306, "MySQL port")
	flag.IntVar(&rowLimit, "l", 10, "Number of rows to display before truncation (default: 10)")
	flag.IntVar(&executionCount, "c", 1, "Number of times to execute the query (default: 1)")
	flag.Parse()

	// Ensure a database name is provided as the first non-flag argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Database name is required.")
		os.Exit(1)
	}
	database := args[0]

	// Validate required parameters
	if *user == "" {
		fmt.Println("Error: MySQL username (-u) is required.")
		os.Exit(1)
	}

	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", *user, *password, *host, *port, database)

	// Open MySQL connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to MySQL server: %v", err)
	}
	fmt.Printf("Connected to MySQL database '%s'!\n", database)

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
		// Generate prompt with HH:MM:SS format
		currentTime := time.Now().Format("15:04:05")
		fmt.Printf("micro-mysql (%s)> ", currentTime)

		// Read user input (full line)
		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		query = strings.TrimSpace(query)

		// Check for exit command
		if isExitCommand(query) {
			fmt.Println("Goodbye!")
			break
		}

		// Handle `SET MICRO` commands
		if handleMicroCommand(query) {
			continue
		}

		// Handle `HELP` command
		if strings.EqualFold(query, "HELP") {
			displayHelp()
			continue
		}

		// Only allow SELECT statements
		if strings.HasPrefix(strings.ToLower(query), "select") {
			executeQuery(db, query, rowLimit, executionCount)
		} else {
			fmt.Println("Only SELECT statements are allowed. Type 'HELP' for available commands.")
		}
	}
}

// isExitCommand checks if the input matches an exit command (case-insensitive)
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

// handleMicroCommand processes SET MICRO commands
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

// parseInt safely converts a string to an integer
func parseInt(s string) (int, error) {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	return value, err
}

// displayHelp prints available commands
func displayHelp() {
	fmt.Println("\nAvailable Commands:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("HELP                 - Display this help message")
	fmt.Println("EXIT, QUIT, \\q, :wq  - Exit the application")
	fmt.Println("SET MICRO COUNT=N    - Set the execution count for queries")
	fmt.Println("SET MICRO LIMIT=N    - Set the maximum number of rows displayed")
	fmt.Println("SELECT ...           - Execute a SELECT query")
	fmt.Println(strings.Repeat("-", 50))
}

func executeQuery(db *sql.DB, query string, rowLimit int, count int) {
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
		columnCount := len(columns)
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if i == 0 {
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println(strings.Join(columns, "\t"))
			fmt.Println(strings.Repeat("-", 50))
		}

		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				fmt.Printf("Row scan error: %v\n", err)
				return
			}

			for i, val := range values {
				if b, ok := val.([]byte); ok {
					values[i] = string(b)
				}
			}

			if rowCount < rowLimit && i == 0 {
				for _, val := range values {
					fmt.Printf("%v\t", val)
				}
				fmt.Println()
			}

			rowCount++
		}

		if rowCount > rowLimit && i == 0 {
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
