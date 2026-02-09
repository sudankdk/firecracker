package main

import (
	"fmt"
)

// StartVM configures and starts a Firecracker VM with optional input drive
func StartVM(sock, kernel, rootfs, inputDrive string) error {
	client := NewClient(sock)

	if err := Put(client, "/machine-config", []byte(`{
		"vcpu_count": 1,
		"mem_size_mib": 512
	}`)); err != nil {
		return err
	}

	if err := Put(client, "/boot-source", []byte(fmt.Sprintf(`{
		"kernel_image_path": "%s",
		"boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
	}`, kernel))); err != nil {
		return err
	}

	if err := Put(client, "/drives/rootfs", []byte(fmt.Sprintf(`{
		"drive_id": "rootfs",
		"path_on_host": "%s",
		"is_root_device": true,
		"is_read_only": false
	}`, rootfs))); err != nil {
		return err
	}

	// Attach secondary input drive if provided (read-only for security)
	if inputDrive != "" {
		if err := Put(client, "/drives/input_drive", []byte(fmt.Sprintf(`{
			"drive_id": "input_drive",
			"path_on_host": "%s",
			"is_root_device": false,
			"is_read_only": true
		}`, inputDrive))); err != nil {
			return fmt.Errorf("failed to attach input drive: %w", err)
		}
	}

	return Put(client, "/actions", []byte(`{
		"action_type": "InstanceStart"
	}`))
}
