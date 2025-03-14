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

const (
  LINE = 60
)

var rowLimit int                                            // TODO: Refactor as local var
var executionCount int                                      // TODO: Refactor as local var

func main() {
    var dbType string

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

	user, password, host, port, database := parseFlags()    // Parse command-line flags
	if password == "" {                                     // Prompt for password if not provided
		password = promptForPassword()
	}

	db := connectToDatabase(dbType, user, password, host, port, database) // Connect to database
	defer db.Close()

	handleSigint(db)                                        // Handle SIGINT (^C)
	handleUserInput(dbType, db)                             // Start the user input loop
}

// Parse command-line flags and return values
func parseFlags() (user, password, host string, port int, database string) {
	userFlag := flag.String("u", "", "Database username")
	passwordFlag := flag.String("p", "", "Database password (optional, will prompt if not provided)")
	hostFlag := flag.String("h", "127.0.0.1", "Database host")
	portFlag := flag.Int("P", getDefaultPort(""), "Database port")
	flag.IntVar(&rowLimit, "l", 10, "Number of rows to display before truncation (default: 10)")
	flag.IntVar(&executionCount, "c", 3, "Number of times to execute the query (default: 3)")
	flag.Parse()

	args := flag.Args()                                     // Ensure a database name is provided
	if len(args) < 1 {
		fmt.Println("Error: Database name is required.")
		os.Exit(1)
	}

	return *userFlag, *passwordFlag, *hostFlag, *portFlag, args[0]
}

func promptForPassword() string {                           // Prompt for password if needed
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
func connectToDatabase(dbType, user, password, host string, port int, database string) *sql.DB {
	dsn := constructDSN(dbType, user, password, host, port, database)
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {                       // Verify connection
		log.Fatalf("Cannot connect to database: %v", err)
	}
	fmt.Printf("Connected to %s database '%s'!\n", dbType, database)

	return db
}

func handleSigint(db *sql.DB) {                             // Handle SIGINT (^C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nExiting due to SIGINT")
		db.Close()
		os.Exit(0)
	}()
}

func handleUserInput(dbType string, db *sql.DB) {           // Handle user input loop
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
			break
		}

		if handleSetCommand(query) {
			continue
		}

		if strings.EqualFold(query, "HELP") {
			displayHelp()
			continue
		}

		if strings.HasPrefix(strings.ToLower(query), "select") {
			executeQuery(db, query, rowLimit, executionCount)
		} else {
			fmt.Println("Only SELECT SQL statements are allowed. Type 'HELP' for available commands.")
		}
	}
}

func executeQuery(db *sql.DB, query string, rowLimit, count int) {
	var totalQueryTime, totalResultTime int64
	var totalRows int

	for iteration := 0; iteration < count; iteration++ {

		queryStart := time.Now().UnixNano()                 // Measure query execution time
		rows, err := db.Query(query)
		queryElapsed := time.Now().UnixNano() - queryStart  // Query execution time

		if err != nil {
			fmt.Printf("Query Error: %v\n", err)
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()                      // Get column names
		if err != nil {
			fmt.Printf("Error fetching columns: %v\n", err)
			return
		}

		columnCount := len(columns)                         // Prepare result storage
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)
		for j := range values {
			valuePtrs[j] = &values[j]
		}

		if iteration == 0 {                                 // Print headers on first execution
            lineSeparator()
			fmt.Println(strings.Join(columns, "\t"))
            lineSeparator()
		}

		rowCount := 0
		renderStart := time.Now().UnixNano()                // Start rendering time
		for rows.Next() {                                   // Iterate over rows 
			err := rows.Scan(valuePtrs...)
			if err != nil {
				fmt.Printf("Row scan error: %v\n", err)
				return
			}

			for j, val := range values {                    // Convert []byte for readable display
				if b, ok := val.([]byte); ok {
					values[j] = string(b)
				}
			}

			if iteration == 0 {                             // Print row data in first iteration only
				if rowCount < rowLimit {
					for _, val := range values {
				fmt.Printf("%v\t", val)
			}
				fmt.Println()
			} else if rowCount == rowLimit {
				fmt.Printf("...")
			}
		}

			rowCount++
		}

		renderElapsed := time.Now().UnixNano() - renderStart // Measure render time

		if iteration == 0 && rowCount > rowLimit {           // notify truncated to `rowLimit` rows
			fmt.Printf(" ... Output truncated at %d rows.\n", rowLimit)
		}

		totalQueryTime += queryElapsed                      // Accumulate total times
		totalResultTime += renderElapsed
		totalRows = rowCount

		// Show query performance details
		fmt.Printf("%d rows (%.3f ms query, %.3f ms result)\n",
			rowCount,
			float64(queryElapsed)/1e6,                      // Convert nanoseconds to milliseconds
			float64(renderElapsed)/1e6,                     // Convert nanoseconds to milliseconds
		)
	}

	if count > 1 {                                          // Print average only for multiple iterations
		fmt.Printf("Average: %d rows (%.3f ms query, %.3f ms result, %d executions)\n",
			totalRows,
			float64(totalQueryTime)/1e6/float64(count),     // Average query execution time
			float64(totalResultTime)/1e6/float64(count),    // Average result processing time
			count,
		)
	}
    lineSeparator()
}

func defaultPortLookup() map[string]int {                   // Define default database ports
        return map[string]int{
                "mysql":      3306,
                "postgresql": 5432,
                "sqlserver":  1433,
                "oracle":     1521,
                "sqlite":     0,
        }
}

func getDefaultPort(dbType string) int {                    // Default port based on database type
    portLookup := defaultPortLookup()
    if port, ok := portLookup[dbType]; ok {
       return port
    }
    return 0
}

func constructDSN(dbType, user, password, host string, port int, database string) string {
	if dbType == "postgresql" {
	    return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable", dbType, user, password, host, port, database)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
}

func isExitCommand(input string) bool {                     // Check exit commands 
	lowerInput := strings.ToLower(strings.TrimSpace(input))
	exitCommands := []string{"exit", "quit", "\\q", ":wq"}  // Be user friendly easter egg

	for _, cmd := range exitCommands {
		if lowerInput == cmd {
			fmt.Println("End of line.")
			return true
		}
	}
	return false
}

func displayHelp() {                                        // Display help
	fmt.Println("\nAvailable Commands:")
    lineSeparator()
	fmt.Printf("HELP                 - Display this message\n")
	fmt.Printf("EXIT                 - End of Line\n")
	fmt.Printf("SET MICRO COUNT=N    - Set number of iterations for queries   (Currently %d)\n", executionCount)
	fmt.Printf("SET MICRO LIMIT=N    - Set rows displayed for first iteration (Currently %d)\n", rowLimit)
	fmt.Printf("SELECT ...           - Execute the given SELECT query\n")
    lineSeparator()
}

func lineSeparator() {
	fmt.Println(strings.Repeat("-", LINE))
}

func handleSetCommand(input string) bool {
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

func parseInt(s string) (int, error) {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	return value, err
}
