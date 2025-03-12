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
	start := time.Now()

	// Execute SELECT query
	rows, err := db.Query(query)
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
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Print column headers
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(strings.Join(columns, "\t"))
	fmt.Println(strings.Repeat("-", 50))

	// Iterate over rows
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			fmt.Printf("Row scan error: %v\n", err)
			return
		}

		// Print row data
		for _, val := range values {
			if val != nil {
				fmt.Printf("%v\t", val)
			} else {
				fmt.Print("NULL\t")
			}
		}
		fmt.Println()
	}

	// Report execution time
	elapsed := time.Since(start)
	fmt.Printf("Query executed in %d µs\n", elapsed.Microseconds())
	fmt.Println(strings.Repeat("-", 50))
}
