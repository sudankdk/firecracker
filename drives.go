package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

const (
	rootDriveName = "root_drive"
	baseDir       = "/mnt/d/firecracker"
	uploadsDir    = baseDir + "/uploads"
	disksDir      = baseDir + "/disks"
	mountBaseDir  = baseDir + "/mnt"
)

type UploadResponse struct {
	JobID  string `json:"jobID"`
	Hash   string `json:"hash"`
	Status string `json:"status"`
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer file.Close()

	jobID := uuid.New().String()
	dstPath := uploadsDir + "/" + jobID + ".bin"

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Failed to create upload file: "+err.Error(), 500)
		return
	}
	defer dst.Close()

	// Save actual file content (not stub)
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), 500)
		return
	}

	log.Printf("File uploaded: %s (jobID: %s, size: %d bytes)", header.Filename, jobID, header.Size)

	// Hash the uploaded file
	hash, err := hashFile(dstPath)
	if err != nil {
		http.Error(w, "Failed to hash file: "+err.Error(), 500)
		return
	}

	log.Printf("File hash (SHA-256): %s", hash)

	// Process the uploaded file (create disk, mount, copy)
	if err := ProcessUploadedFile(jobID); err != nil {
		http.Error(w, "Failed to process file: "+err.Error(), 500)
		return
	}

	diskPath := disksDir + "/input-" + jobID + ".ext4"

	// Run YARA scan automatically
	log.Printf("Starting YARA scan for job %s", jobID)
	yaraScanResult, err := ScanFileWithYara(jobID)
	if err != nil {
		log.Printf("YARA scan error (continuing anyway): %v", err)
	}

	scanStatus := "pending"
	if yaraScanResult != nil {
		scanStatus = yaraScanResult.Status
		if yaraScanResult.MatchCount > 0 {
			scanStatus = fmt.Sprintf("%s (%d detections)", yaraScanResult.Status, yaraScanResult.MatchCount)
			log.Printf("⚠️  MALWARE DETECTED: %d YARA rules matched", yaraScanResult.MatchCount)
			for _, det := range yaraScanResult.Detections {
				log.Printf("   - %s: %s (severity: %s)", det.RuleName, det.Description, det.Severity)
			}
		} else {
			log.Printf("✓ File is clean (no YARA matches)")
		}
	}

	// Store job information in global tracker
	jobs[jobID] = &JobStatus{
		ID:         jobID,
		Hash:       hash,
		UploadTime: fmt.Sprintf("%d", header.Size), // Store file size temporarily
		DiskPath:   diskPath,
		VMStatus:   "ready",
		ScanResult: scanStatus,
	}

	// Return JSON response with job details
	response := UploadResponse{
		JobID:  jobID,
		Hash:   hash,
		Status: "disk_created",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createDiskImage(jobID string) error {
	diskPath := disksDir + "/input-" + jobID + ".ext4"

	// Allocate 50MB disk file using dd (WSL2 compatible)
	// dd is more reliable on Windows-mounted filesystems than fallocate
	cmd := exec.Command("dd", "if=/dev/zero", "of="+diskPath, "bs=1M", "count=50", "status=none")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to allocate disk: %w", err)
	}

	// Create ext4 filesystem
	cmd = exec.Command("mkfs.ext4", "-F", diskPath)
	cmd.Stdout = nil // Suppress mkfs.ext4 verbose output
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create ext4 filesystem: %w", err)
	}

	log.Printf("Created disk image: %s", diskPath)
	return nil
}

func mountDiskImage(jobID, diskPath string) error {
	mountDir := mountBaseDir + "/input-" + jobID

	// Create mount directory with restrictive permissions
	if err := os.MkdirAll(mountDir, 0700); err != nil {
		return fmt.Errorf("failed to create mount directory: %w", err)
	}

	// Mount the disk image
	cmd := exec.Command("mount", "-o", "loop", diskPath, mountDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount disk: %w", err)
	}

	log.Printf("Mounted disk at: %s", mountDir)
	return nil
}

func uploadFileToDrive(jobID string) error {
	src := uploadsDir + "/" + jobID + ".bin"
	mountDir := mountBaseDir + "/input-" + jobID
	dst := mountDir + "/input.bin"

	// Copy uploaded file to mounted disk
	cmd := exec.Command("cp", src, dst)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy file to disk: %w", err)
	}

	// Sync to ensure data is written
	cmd = exec.Command("sync")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	// Unmount the disk
	cmd = exec.Command("umount", mountDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount disk: %w", err)
	}

	log.Printf("File copied to disk and unmounted: %s", dst)
	return nil
}

// ProcessUploadedFile orchestrates the complete pipeline
func ProcessUploadedFile(jobID string) error {
	diskPath := disksDir + "/input-" + jobID + ".ext4"

	// Step 1: Create disk image
	if err := createDiskImage(jobID); err != nil {
		return err
	}

	// Step 2: Mount disk image
	if err := mountDiskImage(jobID, diskPath); err != nil {
		return err
	}

	// Step 3: Copy file to disk and unmount
	if err := uploadFileToDrive(jobID); err != nil {
		return err
	}

	return nil
}
