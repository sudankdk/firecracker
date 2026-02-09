package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// JobStatus tracks the status of an upload job
type JobStatus struct {
	ID         string `json:"id"`
	Hash       string `json:"hash"`
	UploadTime string `json:"uploadTime"`
	DiskPath   string `json:"diskPath"`
	VMStatus   string `json:"vmStatus"`
	ScanResult string `json:"scanResult"`
}

// Simple in-memory job tracker (use database in production)
var jobs = make(map[string]*JobStatus)

func initDirectories() error {
	dirs := []string{
		"/mnt/d/firecracker/uploads",
		"/mnt/d/firecracker/disks",
		"/mnt/d/firecracker/mnt",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("Initialized directory: %s", dir)
	}

	return nil
}

func jobStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract jobID from URL path: /jobs/{jobID}
	path := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID := path

	if jobID == "" {
		http.Error(w, "Job ID required", 400)
		return
	}

	job, exists := jobs[jobID]
	if !exists {
		http.Error(w, "Job not found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func vmScanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	// Extract jobID from URL path: /vm/scan/{jobID}
	path := strings.TrimPrefix(r.URL.Path, "/vm/scan/")
	jobID := path

	if jobID == "" {
		http.Error(w, "Job ID required", 400)
		return
	}

	// Check if job exists
	job, exists := jobs[jobID]
	if !exists {
		http.Error(w, "Job not found", 404)
		return
	}

	// Construct paths for VM
	sock := fmt.Sprintf("/tmp/firecracker-%s.sock", jobID)
	kernel := "/mnt/d/firecracker/hello-vmlinux.bin"
	rootfs := "/mnt/d/firecracker/hello-rootfs.ext4"
	inputDrive := fmt.Sprintf("/mnt/d/firecracker/disks/input-%s.ext4", jobID)

	// Verify input drive exists
	if _, err := os.Stat(inputDrive); os.IsNotExist(err) {
		http.Error(w, "Disk image not found for job", 404)
		return
	}

	log.Printf("Starting VM for job %s with input drive: %s", jobID, inputDrive)

	// Start VM with input drive attached
	if err := StartVM(sock, kernel, rootfs, inputDrive); err != nil {
		http.Error(w, "Failed to start VM: "+err.Error(), 500)
		job.VMStatus = "failed"
		return
	}

	// Update job status
	job.VMStatus = "running"
	job.ScanResult = "scanning..."

	log.Printf("VM started successfully for job %s", jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "vm_started",
		"jobID":  jobID,
		"socket": sock,
	})
}

func main() {
	// Initialize required directories
	if err := initDirectories(); err != nil {
		log.Fatalf("Failed to initialize directories: %v", err)
	}

	// Register handlers
	http.HandleFunc("/upload", uploadHandler)   // From drives.go
	http.HandleFunc("/vm/scan/", vmScanHandler) // VM launch endpoint
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			cleanupHandler(w, r)
		} else if r.Method == "GET" {
			jobStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", 405)
		}
	})
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		// List all jobs
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	})

	// Legacy VM start endpoint (for testing)
	http.HandleFunc("/vm/start", func(w http.ResponseWriter, r *http.Request) {
		err := StartVM(
			"/tmp/firecracker.sock",
			"/mnt/d/firecracker/hello-vmlinux.bin",
			"/mnt/d/firecracker/hello-rootfs.ext4",
			"", // No input drive
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("VM started"))
	})

	port := "8080"
	log.Printf("HTTP server listening on port %s...", port)
	log.Printf("Endpoints:")
	log.Printf("  POST   /upload              - Upload file for scanning")
	log.Printf("  POST   /vm/scan/{jobID}     - Start VM to scan uploaded file")
	log.Printf("  GET    /jobs/{jobID}        - Get job status")
	log.Printf("  GET    /jobs                - List all jobs")
	log.Printf("  DELETE /jobs/{jobID}        - Cleanup job resources")
	log.Printf("  POST   /vm/start            - Start VM (legacy)")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
