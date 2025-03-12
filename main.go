package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Define command-line flags
	user := flag.String("u", "", "MySQL username")
	password := flag.String("p", "", "MySQL password")
	host := flag.String("h", "127.0.0.1", "MySQL host")
	port := flag.Int("P", 3306, "MySQL port")
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

		query = strings.TrimSpace(query) // Remove newlines and spaces

		// Exit conditions
		if query == "exit" || query == "quit" || query == "\\q" || query == ":wq" {
			fmt.Println("Goodbye!")
			break
		}

		// Only allow SELECT statements
		if strings.HasPrefix(strings.ToLower(query), "select") {
			executeQuery(db, query)
		} else {
			fmt.Println("Only SELECT statements are allowed. Type 'exit' to quit.")
		}
	}
}

func executeQuery(db *sql.DB, query string) {
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

	// Print column headers
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(strings.Join(columns, "\t"))
	fmt.Println(strings.Repeat("-", 50))

	// Prepare result storage
	rowCount := 0
	columnCount := len(columns)
	values := make([]interface{}, columnCount)
	valuePtrs := make([]interface{}, columnCount)

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Iterate over rows, limit display to 10 rows but process all
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

		// Print only the first 10 rows
		if rowCount < 10 {
			for _, val := range values {
				fmt.Printf("%v\t", val)
			}
			fmt.Println()
		}

		rowCount++
	}

	// If more than 10 rows, indicate truncation
	if rowCount > 10 {
		fmt.Println("[...] Output truncated at 10 rows.")
	}

	// Measure total execution time (query + result reading)
	totalElapsed := time.Since(totalStart)

	// Output query performance data
	fmt.Printf("%d rows (%.6f ms query, %.6f ms result)\n",
		rowCount,
		float64(queryElapsed.Nanoseconds())/1e6,  // Convert to milliseconds
		float64(totalElapsed.Nanoseconds())/1e6, // Convert to milliseconds
	)
	fmt.Println(strings.Repeat("-", 50))
}
