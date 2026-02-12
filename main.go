package main

import (
	"log"
	"net/http"
	"os"

	handler "github.com/sudankdk/firecracker/internal/Handler"
	"github.com/sudankdk/firecracker/internal/sandboxing"
)

// JobStatus is an alias for Job (for backwards compatibility)
// type JobStatus = Job

// func initDirectories() error {
// 	dirs := []string{
// 		"/mnt/d/firecracker/uploads",
// 		"/mnt/d/firecracker/disks",
// 		"/mnt/d/firecracker/mnt",
// 	}

// 	for _, dir := range dirs {
// 		if err := os.MkdirAll(dir, 0755); err != nil {
// 			return fmt.Errorf("failed to create directory %s: %w", dir, err)
// 		}
// 		log.Printf("Initialized directory: %s", dir)
// 	}

// 	// Initialize YARA environment
// 	if err := PrepareYaraEnvironment(); err != nil {
// 		log.Printf("Warning: YARA initialization failed: %v", err)
// 	}

// 	// Initialize database
// 	if err := InitDatabase(); err != nil {
// 		return fmt.Errorf("failed to initialize database: %w", err)
// 	}

// 	return nil
// }

// func jobStatusHandler(w http.ResponseWriter, r *http.Request) {
// 	// Extract jobID from URL path: /jobs/{jobID}
// 	path := strings.TrimPrefix(r.URL.Path, "/jobs/")
// 	jobID := path

// 	if jobID == "" {
// 		http.Error(w, "Job ID required", 400)
// 		return
// 	}

// 	job, err := GetJob(jobID)
// 	if err != nil {
// 		http.Error(w, "Job not found", 404)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(job)
// }

// func vmScanHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != "POST" {
// 		http.Error(w, "Method not allowed", 405)
// 		return
// 	}

// 	// Extract jobID from URL path: /vm/scan/{jobID}
// 	path := strings.TrimPrefix(r.URL.Path, "/vm/scan/")
// 	jobID := path

// 	if jobID == "" {
// 		http.Error(w, "Job ID required", 400)
// 		return
// 	}

// 	// Check if job exists
// 	job, err := GetJob(jobID)
// 	if err != nil {
// 		http.Error(w, "Job not found", 404)
// 		return
// 	}

// 	// Construct paths for VM
// 	sock := fmt.Sprintf("/tmp/firecracker-%s.sock", jobID)
// 	kernel := "/mnt/d/firecracker/hello-vmlinux.bin"
// 	rootfs := "/mnt/d/firecracker/hello-rootfs.ext4"
// 	inputDrive := fmt.Sprintf("/mnt/d/firecracker/disks/input-%s.ext4", jobID)

// 	// Verify input drive exists
// 	if _, err := os.Stat(inputDrive); os.IsNotExist(err) {
// 		http.Error(w, "Disk image not found for job", 404)
// 		return
// 	}

// 	log.Printf("Starting VM for job %s with input drive: %s", jobID, inputDrive)

// 	// Start VM with input drive attached
// 	if err := StartVM(sock, kernel, rootfs, inputDrive); err != nil {
// 		http.Error(w, "Failed to start VM: "+err.Error(), 500)
// 		UpdateJobStatus(jobID, "failed", job.ScanResult)
// 		return
// 	}

// 	// Update job status
// 	UpdateJobStatus(jobID, "running", "scanning...")

// 	log.Printf("VM started successfully for job %s", jobID)

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"status": "vm_started",
// 		"jobID":  jobID,
// 		"socket": sock,
// 	})
// }

func main() {
	// Initialize required directories
	// if err := initDirectories(); err != nil {
	// 	log.Fatalf("Failed to initialize directories: %v", err)
	// }

	// Register handlers
	// http.HandleFunc("/upload", uploadHandler)   // From drives.go
	// http.HandleFunc("/vm/scan/", vmScanHandler) // VM launch endpoint
	// http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
	// 	switch r.Method {
	// 	case "DELETE":
	// 		cleanupHandler(w, r)
	// 	case "GET":
	// 		jobStatusHandler(w, r)
	// 	default:
	// 		http.Error(w, "Method not allowed", 405)
	// 	}
	// })
	// http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
	// 	// List all jobs
	// 	jobs, err := GetAllJobs()
	// 	if err != nil {
	// 		http.Error(w, "Failed to retrieve jobs", 500)
	// 		return
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(jobs)
	// })

	// // Legacy VM start endpoint (for testing)
	// http.HandleFunc("/vm/start", func(w http.ResponseWriter, r *http.Request) {
	// 	err := StartVM(
	// 		"/tmp/firecracker.sock",
	// 		"/mnt/d/firecracker/hello-vmlinux.bin",
	// 		"/mnt/d/firecracker/hello-rootfs.ext4",
	// 		"", // No input drive
	// 	)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), 500)
	// 		return
	// 	}
	// 	w.Write([]byte("VM started"))
	// })

	// // New scan endpoint
	// http.HandleFunc("/scan/", func(w http.ResponseWriter, r *http.Request) {
	// 	path := strings.TrimPrefix(r.URL.Path, "/scan/")
	// 	jobID := path

	// 	if jobID == "" {
	// 		http.Error(w, "Job ID required", 400)
	// 		return
	// 	}

	// 	// Get existing results if available
	// 	if r.Method == "GET" {
	// 		result, err := GetYaraResults(jobID)
	// 		if err != nil {
	// 			http.Error(w, "Scan results not found", 404)
	// 			return
	// 		}
	// 		w.Header().Set("Content-Type", "application/json")
	// 		json.NewEncoder(w).Encode(result)
	// 		return
	// 	}

	// 	//POST - trigger new scan
	// 	if r.Method == "POST" {
	// 		result, err := ScanFileWithYara(jobID)
	// 		if err != nil {
	// 			log.Printf("YARA scan error: %v", err)
	// 		}

	// 		// Update job status in database
	// 		if result != nil {
	// 			scanStatus := result.Status
	// 			if result.MatchCount > 0 {
	// 				scanStatus = fmt.Sprintf("%s (%d detections)", result.Status, result.MatchCount)
	// 			}
	// 			job, _ := GetJob(jobID)
	// 			if job != nil {
	// 				UpdateJobStatus(jobID, job.VMStatus, scanStatus)
	// 			}
	// 		}

	// 		w.Header().Set("Content-Type", "application/json")
	// 		json.NewEncoder(w).Encode(result)
	// 		return
	// 	}

	// 	http.Error(w, "Method not allowed", 405)
	// })

	// // Stats endpoint
	// http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
	// 	stats, err := GetJobStats()
	// 	if err != nil {
	// 		http.Error(w, "Failed to get stats", 500)
	// 		return
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(stats)
	// })

	if err := os.MkdirAll("/tmp/uploads", 0755); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll("/tmp/vms", 0755); err != nil {
		log.Fatal(err)
	}

	vmManager := &sandboxing.VMManager{
		BaseChrootDir:   "/tmp/vms",
		BaseUploadDir:   "/tmp/uploads",
		KernelPath:      "/mnt/d/firecracker/hello-vmlinux.bin",
		RootfsPath:      "/mnt/d/firecracker/hello-rootfs.ext4",
		JailerPath:      "/mnt/d/firecracker/release-v1.7.0-x86_64/jailer-v1.7.0-x86_64",
		FirecrackerPath: "/mnt/d/firecracker/release-v1.7.0-x86_64/firecracker-v1.7.0-x86_64",
		// BaseChrootDir:   "/srv/vms",
		// BaseUploadDir:   "/srv/uploads",
		// KernelPath:      "/mnt/d/firecracker/hello-vmlinux.bin",
		// RootfsPath:      "/mnt/d/firecracker/hello-rootfs.ext4",
		// JailerPath:      "/opt/firecracker/jailer",
		// FirecrackerPath: "/opt/firecracker/firecracker",
	}

	uploadHandler := &handler.UploadHandler{
		VM: vmManager,
	}

	http.Handle("/upload", uploadHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
