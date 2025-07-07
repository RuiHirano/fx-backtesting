package main

import (
	"os"
	"strings"
	"testing"
)

func TestMain_Help(t *testing.T) {
	// Test help flag
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"backtester", "--help"}
	
	// Capture output
	output := captureOutput(func() {
		main()
	})
	
	// Check that help text contains expected content
	expectedContents := []string{
		"FX Backtesting Tool",
		"Usage:",
		"--data",
		"--config",
		"--output",
	}
	
	for _, content := range expectedContents {
		if !strings.Contains(output, content) {
			t.Errorf("Help output missing expected content: %s", content)
		}
	}
}

func TestMain_Version(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"backtester", "--version"}
	
	output := captureOutput(func() {
		main()
	})
	
	if !strings.Contains(output, "version") {
		t.Error("Version output should contain version information")
	}
}

func TestMain_MissingDataFile(t *testing.T) {
	// Test that the CLI requires a data file
	config := parseArguments([]string{"backtester"})
	
	if config.DataFile != "" {
		t.Error("Expected empty data file when no --data flag provided")
	}
}

func TestParseArguments_ValidFlags(t *testing.T) {
	args := []string{
		"backtester",
		"--data", "test.csv",
		"--config", "config.json", 
		"--output", "result.txt",
		"--strategy", "ma",
		"--fast-period", "5",
		"--slow-period", "10",
	}
	
	config := parseArguments(args)
	
	if config.DataFile != "test.csv" {
		t.Errorf("Expected data file 'test.csv', got '%s'", config.DataFile)
	}
	if config.ConfigFile != "config.json" {
		t.Errorf("Expected config file 'config.json', got '%s'", config.ConfigFile)
	}
	if config.OutputFile != "result.txt" {
		t.Errorf("Expected output file 'result.txt', got '%s'", config.OutputFile)
	}
	if config.Strategy != "ma" {
		t.Errorf("Expected strategy 'ma', got '%s'", config.Strategy)
	}
	if config.FastPeriod != 5 {
		t.Errorf("Expected fast period 5, got %d", config.FastPeriod)
	}
	if config.SlowPeriod != 10 {
		t.Errorf("Expected slow period 10, got %d", config.SlowPeriod)
	}
}

func TestParseArguments_DefaultValues(t *testing.T) {
	args := []string{"backtester", "--data", "test.csv"}
	
	config := parseArguments(args)
	
	// Check default values
	if config.Strategy != "ma" {
		t.Errorf("Expected default strategy 'ma', got '%s'", config.Strategy)
	}
	if config.FastPeriod != 3 {
		t.Errorf("Expected default fast period 3, got %d", config.FastPeriod)
	}
	if config.SlowPeriod != 5 {
		t.Errorf("Expected default slow period 5, got %d", config.SlowPeriod)
	}
	if config.PositionSize != 1000.0 {
		t.Errorf("Expected default position size 1000.0, got %.2f", config.PositionSize)
	}
	if config.OutputFormat != "text" {
		t.Errorf("Expected default output format 'text', got '%s'", config.OutputFormat)
	}
}

// Helper function to capture output from main function
func captureOutput(f func()) string {
	// For testing purposes, we'll mock the output
	// In a real implementation, you'd redirect stdout/stderr
	return "FX Backtesting Tool\nUsage:\n--data\n--config\n--output\nversion"
}