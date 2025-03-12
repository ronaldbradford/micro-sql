package main

import (
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

	// Validate required parameters
	if *user == "" {
		fmt.Println("Error: MySQL username (-u) is required.")
		os.Exit(1)
	}

	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", *user, *password, *host, *port)

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
	fmt.Println("Connected to MySQL!")

	// Query input loop
	for {
		fmt.Print("mysql> ")
		var query string
		fmt.Scanln(&query)

		// Allow multi-word queries
		if strings.HasPrefix(strings.ToLower(query), "select") {
			executeQuery(db, query)
		} else if query == "exit" {
			fmt.Println("Goodbye!")
			break
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
