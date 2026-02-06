package main

import (
	"fmt"
)

func StartVM(sock, kernel, rootfs string) error {
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

	return Put(client, "/actions", []byte(`{
		"action_type": "InstanceStart"
	}`))
}
