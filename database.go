package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Job represents a file upload and scan job in the database
type Job struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	Hash       string    `gorm:"index" json:"hash"`
	FileName   string    `json:"fileName"`
	FileSize   int64     `json:"fileSize"`
	DiskPath   string    `json:"diskPath"`
	VMStatus   string    `json:"vmStatus"`
	ScanResult string    `json:"scanResult"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// InitDatabase initializes the SQLite database with GORM
func InitDatabase() error {
	var err error

	dbPath := "/mnt/d/firecracker/firecracker.db"

	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce log noise
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Job{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Printf("Database initialized: %s", dbPath)
	return nil
}

// CreateJob creates a new job in the database
func CreateJob(job *Job) error {
	result := db.Create(job)
	if result.Error != nil {
		return fmt.Errorf("failed to create job: %w", result.Error)
	}
	return nil
}

// GetJob retrieves a job by ID
func GetJob(jobID string) (*Job, error) {
	var job Job
	result := db.First(&job, "id = ?", jobID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &job, nil
}

// UpdateJob updates an existing job
func UpdateJob(job *Job) error {
	result := db.Save(job)
	if result.Error != nil {
		return fmt.Errorf("failed to update job: %w", result.Error)
	}
	return nil
}

// UpdateJobStatus updates specific fields of a job
func UpdateJobStatus(jobID, vmStatus, scanResult string) error {
	result := db.Model(&Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"vm_status":   vmStatus,
		"scan_result": scanResult,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update job status: %w", result.Error)
	}
	return nil
}

// GetAllJobs retrieves all jobs from the database
func GetAllJobs() ([]Job, error) {
	var jobs []Job
	result := db.Order("created_at DESC").Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobs, nil
}

// DeleteJob removes a job from the database
func DeleteJob(jobID string) error {
	result := db.Delete(&Job{}, "id = ?", jobID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete job: %w", result.Error)
	}
	return nil
}

// GetJobsByStatus retrieves jobs filtered by status
func GetJobsByStatus(status string) ([]Job, error) {
	var jobs []Job
	result := db.Where("vm_status = ?", status).Order("created_at DESC").Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobs, nil
}

// GetJobsByHash retrieves jobs with a specific hash (detect duplicates)
func GetJobsByHash(hash string) ([]Job, error) {
	var jobs []Job
	result := db.Where("hash = ?", hash).Order("created_at DESC").Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobs, nil
}

// GetRecentJobs retrieves the N most recent jobs
func GetRecentJobs(limit int) ([]Job, error) {
	var jobs []Job
	result := db.Order("created_at DESC").Limit(limit).Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobs, nil
}

// JobStats returns statistics about jobs
type JobStats struct {
	Total           int64 `json:"total"`
	Clean           int64 `json:"clean"`
	MalwareDetected int64 `json:"malwareDetected"`
	Running         int64 `json:"running"`
	Failed          int64 `json:"failed"`
}

// GetJobStats returns database statistics
func GetJobStats() (*JobStats, error) {
	stats := &JobStats{}

	// Total jobs
	db.Model(&Job{}).Count(&stats.Total)

	// Clean files
	db.Model(&Job{}).Where("scan_result = ?", "clean").Count(&stats.Clean)

	// Malware detected
	db.Model(&Job{}).Where("scan_result LIKE ?", "%malware%").Count(&stats.MalwareDetected)

	// Running VMs
	db.Model(&Job{}).Where("vm_status = ?", "running").Count(&stats.Running)

	// Failed jobs
	db.Model(&Job{}).Where("vm_status = ?", "failed").Count(&stats.Failed)

	return stats, nil
}
