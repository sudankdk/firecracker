package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// StopVM sends shutdown signal to Firecracker VM
func StopVM(jobID string) error {
	sock := fmt.Sprintf("/tmp/firecracker-%s.sock", jobID)
	client := NewClient(sock)

	// Send InstanceShutdown action
	if err := Put(client, "/actions", []byte(`{
		"action_type": "SendCtrlAltDel"
	}`)); err != nil {
		return fmt.Errorf("failed to send shutdown signal: %w", err)
	}

	log.Printf("Shutdown signal sent to VM for job %s", jobID)
	return nil
}

// CleanupJob removes all resources associated with a job
func CleanupJob(jobID string) error {
	var errors []string

	// 1. Stop VM if running
	if err := StopVM(jobID); err != nil {
		errors = append(errors, fmt.Sprintf("VM stop: %v", err))
	}

	// 2. Remove uploaded file
	uploadPath := uploadsDir + "/" + jobID + ".bin"
	if err := os.Remove(uploadPath); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("upload removal: %v", err))
	} else {
		log.Printf("Removed upload file: %s", uploadPath)
	}

	// 3. Check if disk is mounted and unmount
	mountDir := mountBaseDir + "/input-" + jobID
	cmd := exec.Command("umount", mountDir)
	if err := cmd.Run(); err != nil {
		// Ignore error if not mounted
		log.Printf("Unmount attempt (may not be mounted): %s", mountDir)
	}

	// 4. Remove mount directory
	if err := os.Remove(mountDir); err != nil && !os.IsNotExist(err) {
		log.Printf("Mount directory removal (may not exist): %v", err)
	}

	// 5. Remove disk image
	diskPath := disksDir + "/input-" + jobID + ".ext4"
	if err := os.Remove(diskPath); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("disk removal: %v", err))
	} else {
		log.Printf("Removed disk image: %s", diskPath)
	}

	// 6. Remove Firecracker socket
	sockPath := fmt.Sprintf("/tmp/firecracker-%s.sock", jobID)
	if err := os.Remove(sockPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Socket removal (may not exist): %v", err)
	}

	// 7. Remove job from database
	if err := DeleteJob(jobID); err != nil {
		errors = append(errors, fmt.Sprintf("database deletion: %v", err))
	} else {
		log.Printf("Removed job from database: %s", jobID)
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup had errors: %s", strings.Join(errors, "; "))
	}

	log.Printf("Successfully cleaned up job: %s", jobID)
	return nil
}

// cleanupHandler handles DELETE /jobs/{jobID}
func cleanupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	// Extract jobID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID := path

	if jobID == "" {
		http.Error(w, "Job ID required", 400)
		return
	}

	// Check if job exists
	_, err := GetJob(jobID)
	if err != nil {
		http.Error(w, "Job not found", 404)
		return
	}

	// Perform cleanup
	if err := CleanupJob(jobID); err != nil {
		log.Printf("Cleanup error for job %s: %v", jobID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200) // Return 200 even with partial errors
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "partial_cleanup",
			"jobID":   jobID,
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "cleaned",
		"jobID":  jobID,
	})
}
