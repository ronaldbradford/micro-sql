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

	"golang.org/x/term"

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
	case strings.HasSuffix(progName, "psql"):
		dbType = "postgresql"
	case strings.HasSuffix(progName, "mysql"):
		dbType = "mysql"
	default:
		fmt.Println("Error: Please use `micro-mysql` or `micro-psql` to run this program.")
		os.Exit(1)
	}

	// Parse command-line flags
	user, password, host, port, database := parseFlags()

	// Prompt for password if not provided
	if password == "" {
		password = promptForPassword()
	}

	// Connect to database
	db := connectToDatabase(user, password, host, port, database)
	defer db.Close()

	// Handle SIGINT (^C)
	handleSigint(db)

	// Start the user input loop
	handleUserInput(db)
}

// Parse command-line flags and return values
func parseFlags() (user, password, host string, port int, database string) {
	userFlag := flag.String("u", "", "Database username")
	passwordFlag := flag.String("p", "", "Database password (optional, will prompt if not provided)")
	hostFlag := flag.String("h", "127.0.0.1", "Database host")
	portFlag := flag.Int("P", defaultPort(), "Database port")
	flag.IntVar(&rowLimit, "l", 10, "Number of rows to display before truncation (default: 10)")
	flag.IntVar(&executionCount, "c", 1, "Number of times to execute the query (default: 1)")
	flag.Parse()

	// Ensure a database name is provided
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Database name is required.")
		os.Exit(1)
	}

	return *userFlag, *passwordFlag, *hostFlag, *portFlag, args[0]
}

// Prompt for password securely without displaying input
func promptForPassword() string {
	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Move to a new line after password input
	if err != nil {
		fmt.Println("Error reading password.")
		os.Exit(1)
	}
	return string(bytePassword)
}

// Connect to the database and return the connection
func connectToDatabase(user, password, host string, port int, database string) *sql.DB {
	dsn := constructDSN(dbType, user, password, host, port, database)
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	fmt.Printf("Connected to %s database '%s'!\n", dbType, database)

	return db
}

// Handle SIGINT (^C) and clean up database connection
func handleSigint(db *sql.DB) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived SIGINT, exiting...")
		db.Close()
		os.Exit(0)
	}()
}

// Handle user input loop
func handleUserInput(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)

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

// Execute SELECT queries
func executeQuery(db *sql.DB, query string, rowLimit, count int) {
	var totalQueryTime, totalResultTime time.Duration
	var totalRows int

	for i := 0; i < count; i++ {
		totalStart := time.Now()

		// Start query execution timer
		queryStart := time.Now()
		rows, err := db.Query(query)
		queryElapsed := time.Since(queryStart)

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
		columnCount := len(columns)
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)
		for j := range values {
			valuePtrs[j] = &values[j]
		}

		rowCount := 0

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

			// Convert []byte to string for readable output
			for j, val := range values {
				if b, ok := val.([]byte); ok {
					values[j] = string(b)
				}
			}

			// Print actual row data only in the first iteration and up to `rowLimit`
			if i == 0 && rowCount < rowLimit {
				for _, val := range values {
					fmt.Printf("%v\t", val)
				}
				fmt.Println()
			}

			rowCount++
		}

		// If more than `rowLimit` rows, indicate truncation (only in first iteration)
		if i == 0 && rowCount > rowLimit {
			fmt.Printf("[...] Output truncated at %d rows.\n", rowLimit)
		}

		// Measure total execution time (query + result reading)
		totalElapsed := time.Since(totalStart)

		// Accumulate times
		totalQueryTime += queryElapsed
		totalResultTime += totalElapsed
		totalRows = rowCount

		// Show query performance details
		fmt.Printf("%d rows (%.6f ms query, %.6f ms result)\n",
			rowCount,
			float64(queryElapsed.Nanoseconds())/1e6,
			float64(totalElapsed.Nanoseconds())/1e6,
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

// Default port based on database type
func defaultPort() int {
	if dbType == "postgresql" {
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
