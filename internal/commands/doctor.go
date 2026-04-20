package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics",
	Long:  `Run diagnostic checks on the wacli installation and configuration`,
	RunE:  runDoctor,
}

// doctorFlags holds the flags for doctor command
var doctorFlags = struct {
	jsonOutput bool
}{}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFlags.jsonOutput, "json", false, "Output in JSON format")
}

// DiagnosticResult represents a single diagnostic check result
type DiagnosticResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"` // "pass", "fail", "warn"
	Message string `json:"message,omitempty"`
}

// DoctorResult represents the overall diagnostic result
type DoctorResult struct {
	Timestamp   time.Time          `json:"timestamp"`
	Version     string              `json:"version"`
	OS          string              `json:"os"`
	Arch        string              `json:"arch"`
	Checks      []*DiagnosticResult `json:"checks"`
	AllPassed   bool                `json:"all_passed"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	results := make([]*DiagnosticResult, 0)

	// Check 1: Go version
	checkGoVersion := func() {
		out, err := exec.Command("go", "version").Output()
		if err != nil {
			results = append(results, &DiagnosticResult{
				Check:   "go_version",
				Status:  "fail",
				Message: "Go is not installed or not in PATH",
			})
		} else {
			results = append(results, &DiagnosticResult{
				Check:   "go_version",
				Status:  "pass",
				Message: string(out),
			})
		}
	}
	checkGoVersion()

	// Check 2: Database connectivity
	checkDatabase := func() {
		db := GetDB()
		if db == nil {
			results = append(results, &DiagnosticResult{
				Check:   "database",
				Status:  "fail",
				Message: "Database not initialized",
			})
		} else {
			count, err := db.GetMessageCount()
			if err != nil {
				results = append(results, &DiagnosticResult{
					Check:   "database",
					Status:  "fail",
					Message: fmt.Sprintf("Database error: %v", err),
				})
			} else {
				results = append(results, &DiagnosticResult{
					Check:   "database",
					Status:  "pass",
					Message: fmt.Sprintf("Connected. Messages: %d", count),
				})
			}
		}
	}
	checkDatabase()

	// Check 3: WhatsApp client connection
	checkClientConnection := func() {
		waClient := GetClient()
		if waClient == nil {
			results = append(results, &DiagnosticResult{
				Check:   "wa_client",
				Status:  "fail",
				Message: "WhatsApp client not initialized",
			})
		} else if !waClient.IsConnected() {
			results = append(results, &DiagnosticResult{
				Check:   "wa_client",
				Status:  "warn",
				Message: "WhatsApp client not connected. Run 'wacli auth' to connect.",
			})
		} else {
			jid := waClient.GetMyJID()
			results = append(results, &DiagnosticResult{
				Check:   "wa_client",
				Status:  "pass",
				Message: fmt.Sprintf("Connected as %s", jid),
			})
		}
	}
	checkClientConnection()

	// Check 4: Storage directory
	checkStorageDir := func() {
		storePath := os.Getenv("WACLI_STORE")
		if storePath == "" {
			home, _ := os.UserHomeDir()
			storePath = home + "/.wacli"
		}
		if _, err := os.Stat(storePath); os.IsNotExist(err) {
			results = append(results, &DiagnosticResult{
				Check:   "storage_dir",
				Status:  "warn",
				Message: fmt.Sprintf("Storage directory does not exist: %s", storePath),
			})
		} else {
			results = append(results, &DiagnosticResult{
				Check:   "storage_dir",
				Status:  "pass",
				Message: storePath,
			})
		}
	}
	checkStorageDir()

	// Check 5: Required dependencies
	checkDependencies := func() {
		deps := []string{"sqlite3"}
		allFound := true
		missing := ""
		for _, dep := range deps {
			if _, err := exec.LookPath(dep); err != nil {
				allFound = false
				missing += dep + " "
			}
		}
		if !allFound {
			results = append(results, &DiagnosticResult{
				Check:   "dependencies",
				Status:  "warn",
				Message: fmt.Sprintf("Some system dependencies may be missing: %s", missing),
			})
		} else {
			results = append(results, &DiagnosticResult{
				Check:   "dependencies",
				Status:  "pass",
				Message: "All system dependencies found",
			})
		}
	}
	checkDependencies()

	// Check 6: Network connectivity (basic)
	checkNetwork := func() {
		// Check if we can reach WhatsApp servers
		// For now, just check if the client is connected which implies network is working
		waClient := GetClient()
		if waClient != nil && waClient.IsConnected() {
			results = append(results, &DiagnosticResult{
				Check:   "network",
				Status:  "pass",
				Message: "WhatsApp network connection active",
			})
		} else {
			results = append(results, &DiagnosticResult{
				Check:   "network",
				Status:  "warn",
				Message: "Cannot verify network connectivity",
			})
		}
	}
	checkNetwork()

	// Determine overall status
	allPassed := true
	hasWarnings := false
	for _, r := range results {
		if r.Status == "fail" {
			allPassed = false
		} else if r.Status == "warn" {
			hasWarnings = true
		}
	}

	overallResult := DoctorResult{
		Timestamp: time.Now(),
		Version:   "1.0.0",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Checks:    results,
		AllPassed: allPassed,
	}

	if doctorFlags.jsonOutput {
		output, _ := json.MarshalIndent(overallResult, "", "  ")
		fmt.Println(string(output))
	} else {
		printDoctorResults(&overallResult, hasWarnings)
	}

	if !allPassed {
		return fmt.Errorf("diagnostic checks failed")
	}
	return nil
}

func printDoctorResults(result *DoctorResult, hasWarnings bool) {
	fmt.Println("=== wacli Diagnostic Report ===")
	fmt.Printf("Timestamp: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Version: %s\n", result.Version)
	fmt.Printf("Platform: %s/%s\n\n", result.OS, result.Arch)

	fmt.Println("--- Checks ---")
	for _, check := range result.Checks {
		statusIcon := "✓"
		switch check.Status {
		case "pass":
			statusIcon = "✓"
		case "fail":
			statusIcon = "✗"
		case "warn":
			statusIcon = "⚠"
		}
		fmt.Printf("%s [%s] %s\n", statusIcon, check.Check, check.Message)
	}

	fmt.Println()
	if result.AllPassed && !hasWarnings {
		fmt.Println("✓ All checks passed!")
	} else if result.AllPassed && hasWarnings {
		fmt.Println("⚠ All checks passed, but some warnings were reported.")
	} else {
		fmt.Println("✗ Some checks failed. Please review the errors above.")
	}
}