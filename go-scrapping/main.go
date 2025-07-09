package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

// ScrapedItem represents a title and URL scraped from a website
type ScrapedItem struct {
	Title string
	URL   string
}

// Config holds application configuration
type Config struct {
	DBPath      string
	Concurrency int
	Timeout     time.Duration
	UserAgent   string
}

func main() {
	// Parse command line flags
	config := parseFlags()

	// Set up logging
	logger := log.New(os.Stdout, "[scraper] ", log.LstdFlags)

	// Create context that can be canceled on SIGINT
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		logger.Println("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Initialize SQLite DB
	db, err := initDB(config.DBPath)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create table to store titles
	if err := createTable(db); err != nil {
		logger.Fatalf("Failed to create table: %v", err)
	}

	// URLs to scrape - in a real app, these could come from a config file or command line args
	urls := getURLsToScrape()

	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Process URLs with worker pool pattern
	if err := processURLs(ctx, urls, db, client, config, logger); err != nil {
		logger.Fatalf("Error processing URLs: %v", err)
	}

	logger.Println("Scraping completed successfully!")
}

func parseFlags() Config {
	dbPath := flag.String("db", "./scraped_titles.db", "Path to SQLite database file")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent scrapers")
	timeout := flag.Duration("timeout", 30*time.Second, "HTTP request timeout")
	userAgent := flag.String("user-agent", "GoScraper/1.0", "User-Agent for HTTP requests")
	flag.Parse()

	return Config{
		DBPath:      *dbPath,
		Concurrency: *concurrency,
		Timeout:     *timeout,
		UserAgent:   *userAgent,
	}
}

func getURLsToScrape() []string {
	// In a real application, you might load these from a file or API
	baseURL := "https://news.ycombinator.com/news?p="
	urls := make([]string, 0, 30)

	// Add main HN page
	urls = append(urls, "https://news.ycombinator.com/")

	// Add multiple pages of HN
	for i := 2; i <= 30; i++ {
		urls = append(urls, fmt.Sprintf("%s%d", baseURL, i))
	}

	return urls
}

// Initialize the SQLite DB
func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Create a table to store scraped titles
func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS titles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_url ON titles(url);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// Process URLs using a worker pool pattern
func processURLs(ctx context.Context, urls []string, db *sql.DB, client *http.Client, config Config, logger *log.Logger) error {
	totalURLs := len(urls)
	logger.Printf("Starting to process %d URLs with %d workers", totalURLs, config.Concurrency)

	// Create job channel and initialize worker pool
	jobs := make(chan string, totalURLs)
	results := make(chan ScrapedItem, totalURLs*30) // Each page might have multiple items
	errors := make(chan error, totalURLs)

	// Create a new WaitGroup for workers
	var wg sync.WaitGroup

	// Create tx stmt once for reuse
	stmt, err := db.Prepare("INSERT INTO titles (title, url) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Start the worker pool
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			worker(ctx, workerId, jobs, results, errors, client, config, logger)
		}(i)
	}

	// Database writer goroutine
	var dbWg sync.WaitGroup
	dbWg.Add(1)
	go func() {
		defer dbWg.Done()
		processResults(ctx, results, stmt, logger)
	}()

	// Error handler goroutine
	var errWg sync.WaitGroup
	errWg.Add(1)
	go func() {
		defer errWg.Done()
		for err := range errors {
			logger.Printf("Error: %v", err)
		}
	}()

	// Send jobs to workers
	go func() {
		for _, url := range urls {
			select {
			case <-ctx.Done():
				break
			case jobs <- url:
				// Job sent
			}
		}
		close(jobs)
	}()

	// Wait for all workers to complete
	wg.Wait()
	close(results)
	close(errors)

	// Wait for the DB writer to finish
	dbWg.Wait()
	errWg.Wait()

	return nil
}

// Worker processes URLs from the jobs channel
func worker(ctx context.Context, id int, jobs <-chan string, results chan<- ScrapedItem, errors chan<- error, client *http.Client, config Config, logger *log.Logger) {
	for url := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			logger.Printf("Worker %d processing %s", id, url)
			items, err := scrapeURL(ctx, url, client, config)
			if err != nil {
				errors <- fmt.Errorf("worker %d failed to scrape %s: %w", id, url, err)
				continue
			}

			// Send all scraped items to results channel
			for _, item := range items {
				select {
				case <-ctx.Done():
					return
				case results <- item:
					// Result sent
				}
			}
		}
	}
}

// Process results and save to database
func processResults(ctx context.Context, results <-chan ScrapedItem, stmt *sql.Stmt, logger *log.Logger) {
	count := 0
	for item := range results {
		select {
		case <-ctx.Done():
			logger.Println("Context canceled, stopping result processing")
			return
		default:
			if err := insertTitle(stmt, item); err != nil {
				logger.Printf("Failed to insert item: %v", err)
			} else {
				count++
				if count%100 == 0 {
					logger.Printf("Processed %d items so far", count)
				}
			}
		}
	}
	logger.Printf("Total items saved to database: %d", count)
}

// Scrape a URL for titles
func scrapeURL(ctx context.Context, url string, client *http.Client, config Config) ([]ScrapedItem, error) {
	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var items []ScrapedItem

	// For Hacker News, find story titles (adjust selector as needed)
	doc.Find(".titleline > a").Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		href, exists := s.Attr("href")

		if exists && title != "" {
			item := ScrapedItem{
				Title: title,
				URL:   href,
			}
			items = append(items, item)
		}
	})

	if len(items) == 0 {
		// Fallback to generic title if no stories found
		title := doc.Find("title").Text()
		if title != "" {
			items = append(items, ScrapedItem{
				Title: title,
				URL:   url,
			})
		}
	}

	return items, nil
}

// Insert a scraped title into the SQLite DB
func insertTitle(stmt *sql.Stmt, item ScrapedItem) error {
	_, err := stmt.Exec(item.Title, item.URL)
	if err != nil {
		return fmt.Errorf("failed to insert title: %w", err)
	}
	return nil
}
