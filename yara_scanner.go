package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	yaraRulesDir  = "/mnt/d/firecracker/yara_rules"
	yaraOutputDir = "/mnt/d/firecracker/scan_results"
)

type Detection struct {
	RuleName    string   `json:"ruleName"`
	Tags        []string `json:"tags"`
	Description string   `json:"description,omitempty"`
	Severity    string   `json:"severity,omitempty"`
}

type YaraScanResult struct {
	JobID      string      `json:"jobID"`
	Timestamp  string      `json:"timestamp"`
	Detections []Detection `json:"detections"`
	TotalRules int         `json:"totalRules"`
	MatchCount int         `json:"matchCount"`
	Status     string      `json:"status"`
	ErrorMsg   string      `json:"errorMsg,omitempty"`
	ScanTime   float64     `json:"scanTime"`
}

func PrepareYaraEnvironment() error {
	dirs := []string{yaraRulesDir, yaraOutputDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	defaultRulesPath := filepath.Join(yaraRulesDir, "default_rules.yar")
	if _, err := os.Stat(defaultRulesPath); os.IsNotExist(err) {
		if err := createDefaultYaraRules(defaultRulesPath); err != nil {
			return err
		}
		log.Printf("Created default YARA rules at: %s", defaultRulesPath)
	}
	return nil
}
func createDefaultYaraRules(path string) error {
	rules := `rule Suspicious_EXE_Header {
    meta:
        description = "Detects Windows executable files"
        severity = "info"
    strings:
        $mz = "MZ"
    condition:
        $mz at 0
}

rule Potential_Ransomware_Keywords {
    meta:
        description = "Contains ransomware-related keywords"
        severity = "high"
    strings:
        $ransom1 = "encrypted" nocase
        $ransom2 = "bitcoin" nocase
        $ransom3 = "payment" nocase
        $ransom4 = "decrypt" nocase
    condition:
        3 of them
}

rule Suspicious_Shell_Commands {
    meta:
        description = "Shell command execution patterns"
        severity = "medium"
    strings:
        $cmd1 = "cmd.exe" nocase
        $cmd2 = "powershell" nocase
        $exec1 = "exec" nocase
        $exec2 = "system" nocase
    condition:
        any of ($cmd*) and any of ($exec*)
}

rule Crypto_Mining_Indicators {
    meta:
        description = "Cryptocurrency mining patterns"
        severity = "high"
    strings:
        $crypto1 = "monero" nocase
        $crypto2 = "mining" nocase
        $crypto3 = "stratum" nocase
    condition:
        2 of them
}

rule Keylogger_Indicators {
    meta:
        description = "Keylogger behavior patterns"
        severity = "critical"
    strings:
        $key1 = "GetAsyncKeyState" nocase
        $key2 = "keypress" nocase
        $log = "log" nocase
    condition:
        any of ($key*) and $log
}`

	return os.WriteFile(path, []byte(rules), 0644)
}

func ScanFileWithYara(jobID string) (*YaraScanResult, error) {
	startTime := time.Now()
	result := &YaraScanResult{
		JobID:      jobID,
		Timestamp:  time.Now().Format(time.RFC3339),
		Detections: []Detection{},
		Status:     "clean",
	}

	filePath := uploadsDir + "/" + jobID + ".bin"

	if _, err := exec.LookPath("yara"); err != nil {
		result.Status = "error"
		result.ErrorMsg = "YARA not installed. Install: sudo apt-get install yara"
		return result, fmt.Errorf("yara not found: %w", err)
	}

	ruleFiles, err := filepath.Glob(filepath.Join(yaraRulesDir, "*.yar"))
	if err != nil {
		result.Status = "error"
		result.ErrorMsg = fmt.Sprintf("Failed to find YARA rules: %v", err)
		return result, err
	}

	if len(ruleFiles) == 0 {
		result.Status = "error"
		result.ErrorMsg = "No YARA rules found"
		return result, fmt.Errorf("no YARA rules available")
	}

	result.TotalRules = len(ruleFiles)
	log.Printf("Scanning %s with %d YARA rule files", filePath, len(ruleFiles))

	for _, ruleFile := range ruleFiles {
		detections, err := runYaraScan(ruleFile, filePath)
		if err != nil {
			log.Printf("Warning: YARA scan failed: %v", err)
			continue
		}
		result.Detections = append(result.Detections, detections...)
	}

	result.MatchCount = len(result.Detections)
	result.ScanTime = time.Since(startTime).Seconds()

	if result.MatchCount > 0 {
		result.Status = "malware_detected"
		for _, det := range result.Detections {
			if det.Severity == "critical" {
				result.Status = "critical_threat"
				break
			}
		}
	}

	saveYaraResults(jobID, result)
	return result, nil
}

func runYaraScan(ruleFile, targetFile string) ([]Detection, error) {
	var detections []Detection
	cmd := exec.Command("yara", "-m", ruleFile, targetFile)
	output, err := cmd.CombinedOutput()

	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return nil, fmt.Errorf("yara execution failed: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			detections = append(detections, Detection{
				RuleName:    parts[0],
				Tags:        []string{},
				Description: "YARA rule match",
				Severity:    "medium",
			})
		}
	}
	return detections, nil
}

func saveYaraResults(jobID string, result *YaraScanResult) error {
	resultPath := filepath.Join(yaraOutputDir, jobID+".json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(resultPath, data, 0644)
}

func GetYaraResults(jobID string) (*YaraScanResult, error) {
	resultPath := filepath.Join(yaraOutputDir, jobID+".json")
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, err
	}
	var result YaraScanResult
	json.Unmarshal(data, &result)
	return &result, nil
}
