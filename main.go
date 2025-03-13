package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Define command-line flags
	user := flag.String("u", "", "MySQL username")
	password := flag.String("p", "", "MySQL password")
	host := flag.String("h", "127.0.0.1", "MySQL host")
	port := flag.Int("P", 3306, "MySQL port")
	limit := flag.Int("l", 10, "Number of rows to display before truncation (default: 10)")
	count := flag.Int("c", 1, "Number of times to execute the query and iterate over results (default: 1)")
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

	reader := bufio.NewReader(os.Stdin) // Allows full-line input

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

		query = strings.TrimSpace(query) // Preserve case for execution

		// Exit conditions (case-insensitive check)
		if isExitCommand(query) {
			fmt.Println("Goodbye!")
			break
		}

		// Only allow SELECT statements
		if strings.HasPrefix(strings.ToLower(query), "select") {
			executeQuery(db, query, *limit, *count)
		} else {
			fmt.Println("Only SELECT statements are allowed. Type 'exit' to quit.")
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

func executeQuery(db *sql.DB, query string, rowLimit int, count int) {
	var totalQueryTime time.Duration
	var totalResultTime time.Duration
	var totalRows int

	for i := 0; i < count; i++ {
		totalStart := time.Now() // Start total execution timer

		// Start query execution timer
		queryStart := time.Now()
		rows, err := db.Query(query)
		queryElapsed := time.Since(queryStart) // End query execution timer

		if err != nil {
			fmt.Printf("Query Error: %v\n", err)
			return
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			fmt.Printf("Error fetching columns: %v\n", err)
			return
		}

		// Prepare result storage
		rowCount := 0
		columnCount := len(columns)
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Print column headers only on the first execution
		if i == 0 {
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println(strings.Join(columns, "\t"))
			fmt.Println(strings.Repeat("-", 50))
		}

		// Iterate over rows, limit display to `rowLimit` but process all
		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				fmt.Printf("Row scan error: %v\n", err)
				return
			}

			// Convert byte slices to strings
			for i, val := range values {
				if b, ok := val.([]byte); ok {
					values[i] = string(b) // Convert []byte to string
				}
			}

			// Print only up to `rowLimit` rows on the first execution
			if rowCount < rowLimit && i == 0 {
				for _, val := range values {
					fmt.Printf("%v\t", val)
				}
				fmt.Println()
			}

			rowCount++
		}

		// If more than `rowLimit` rows and first execution, indicate truncation
		if rowCount > rowLimit && i == 0 {
			fmt.Printf("[...] Output truncated at %d rows.\n", rowLimit)
		}

		// Measure total execution time (query + result reading)
		totalElapsed := time.Since(totalStart)

		// Accumulate times
		totalQueryTime += queryElapsed
		totalResultTime += totalElapsed
		totalRows = rowCount

		// Print execution result
		fmt.Printf("%d rows (%.6f ms query, %.6f ms result)\n",
			rowCount,
			float64(queryElapsed.Nanoseconds())/1e6,  // Convert to milliseconds
			float64(totalElapsed.Nanoseconds())/1e6, // Convert to milliseconds
		)
	}

	// Print average only if count > 1
	if count > 1 {
		fmt.Printf("Average: %d rows (%.6f ms query, %.6f ms result over %d runs)\n",
			totalRows,
			float64(totalQueryTime.Nanoseconds())/1e6/float64(count),  // Average query execution time
			float64(totalResultTime.Nanoseconds())/1e6/float64(count), // Average total execution time
			count,
		)
	}
	fmt.Println(strings.Repeat("-", 50))
}
