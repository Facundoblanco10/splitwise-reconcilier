package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/facundo/splitwise-reconcilier/internal/mapping"
	"github.com/facundo/splitwise-reconcilier/internal/report"
	"github.com/facundo/splitwise-reconcilier/internal/splitwise"
)

func main() {
	infoLog := log.New(os.Stdout, "[INFO] ", 0)
	errLog := log.New(os.Stderr, "[ERROR] ", 0)

	monthFlag := flag.String("month", "", "Month to process in YYYY-MM format (e.g. 2026-04)")
	configPath := flag.String("config", "config.yaml", "Path to config.yaml")
	flag.Parse()

	if *monthFlag == "" {
		errLog.Fatal("--month flag is required (e.g. --month 2026-04)")
	}

	// Load .env
	if err := godotenv.Load(); err != nil {
		infoLog.Println("No .env file found, reading from environment")
	}

	apiKey := os.Getenv("SPLITWISE_API_KEY")
	if apiKey == "" {
		errLog.Fatal("SPLITWISE_API_KEY is not set")
	}

	// Load config
	cfg, err := mapping.LoadConfig(*configPath)
	if err != nil {
		errLog.Fatalf("Loading config: %v", err)
	}

	// Parse month
	month, err := time.Parse("2006-01", *monthFlag)
	if err != nil {
		errLog.Fatalf("Invalid --month value %q: expected YYYY-MM", *monthFlag)
	}
	from := month
	to := month.AddDate(0, 1, 0).Add(-time.Second) // last second of the month

	// Validate API key
	client := splitwise.NewClient(apiKey)
	infoLog.Println("Validating API key...")
	user, err := client.GetCurrentUser()
	if err != nil {
		errLog.Fatalf("Authenticating with Splitwise: %v", err)
	}
	infoLog.Printf("Authenticated as %s %s (%s)", user.FirstName, user.LastName, user.Email)

	// Fetch group members — used for column headers and card auto-detection
	group, err := client.GetGroup(cfg.Splitwise.GroupID)
	if err != nil {
		errLog.Fatalf("Fetching group members: %v", err)
	}
	myName, partnerName := resolveDisplayNames(group.Members, cfg.Splitwise.MyUserID)
	userFirstNames := buildFirstNameMap(group.Members)

	// Fetch expenses
	infoLog.Printf("Fetching expenses for group %d from %s to %s...",
		cfg.Splitwise.GroupID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	expenses, err := client.GetExpenses(cfg.Splitwise.GroupID, from, to)
	if err != nil {
		errLog.Fatalf("Fetching expenses: %v", err)
	}
	infoLog.Printf("Fetched %d active expenses", len(expenses))

	// Map cards
	mapper := mapping.NewMapper(cfg, userFirstNames)
	mapped := mapper.MapAll(expenses)

	// Tally results
	cardTotals := map[string]float64{}
	unassignedCount := 0
	for _, e := range mapped {
		cardTotals[e.Card] += e.MyShare + e.TheirShare
		if e.Card == mapping.Unassigned {
			unassignedCount++
		}
	}

	// Generate report
	if err := os.MkdirAll("output", 0755); err != nil {
		errLog.Fatalf("Creating output directory: %v", err)
	}
	outputPath := fmt.Sprintf("output/report_%s.xlsx", *monthFlag)
	infoLog.Printf("Generating report at %s...", outputPath)
	if err := report.Generate(mapped, cfg, month, outputPath, myName, partnerName); err != nil {
		errLog.Fatalf("Generating report: %v", err)
	}

	// Summary
	fmt.Println()
	fmt.Printf("=== Summary %s ===\n", *monthFlag)
	fmt.Printf("Total expenses processed: %d\n", len(mapped))
	fmt.Println()
	fmt.Println("Per card (total):")
	for card, total := range cardTotals {
		if card == mapping.Unassigned {
			continue
		}
		fmt.Printf("  %-25s  %.2f\n", mapping.FormatCardName(card), total)
	}
	if unassignedCount > 0 {
		fmt.Printf("\n  Unassigned (%d expenses): %.2f\n", unassignedCount, cardTotals[mapping.Unassigned])
	}
	fmt.Printf("\nReport saved to: %s\n", outputPath)
}

// resolveDisplayNames returns full display names for Excel column headers.
func resolveDisplayNames(members []splitwise.GroupMember, myUserID int) (myName, partnerName string) {
	var others []string
	for _, m := range members {
		name := strings.TrimSpace(m.FirstName + " " + m.LastName)
		if m.ID == myUserID {
			myName = name
		} else {
			others = append(others, name)
		}
	}
	if myName == "" {
		myName = "Me"
	}
	if len(others) == 0 {
		partnerName = "Partner"
	} else {
		partnerName = strings.Join(others, " / ")
	}
	return
}

// buildFirstNameMap returns a map of Splitwise user ID → first name,
// used by the Mapper to auto-complete card tags and detect the payer.
func buildFirstNameMap(members []splitwise.GroupMember) map[int]string {
	m := make(map[int]string, len(members))
	for _, member := range members {
		m[member.ID] = member.FirstName
	}
	return m
}
