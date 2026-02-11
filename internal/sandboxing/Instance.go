package sandboxing

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/sudankdk/firecracker/internal/domain"
	client "github.com/sudankdk/firecracker/internal/httpclient"
)


type VMManager struct {
	BaseChrootDir string // e.g., "/srv/vms"
	BaseUploadDir string // e.g., "/srv/uploads"
	KernelPath    string
	RootfsPath    string
}


// CreateVMMetadata generates unique IDs, socket path, and TAP name
func CreateVMMetadata() (*domain.VM, error) {
	vmID := uuid.New().String()
	apiSock := fmt.Sprintf("/tmp/firecracker/%s.api.sock", vmID)
	tap := fmt.Sprintf("tap-%s", vmID[:8])

	return &domain.VM{
		ID:      vmID,
		APISock: apiSock,
		TapName: tap,
	}, nil
}

// CreateTAP creates a host TAP interface
func CreateTAP(vm *domain.VM) error {
	if err := exec.Command("ip", "tuntap", "add", vm.TapName, "mode", "tap").Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "link", "set", vm.TapName, "up").Run(); err != nil {
		return err
	}
	return nil
}

// RunFirecracker launches Firecracker process asynchronously
func RunFirecracker(vm *domain.VM) error {
	cmd := exec.Command(
		"./firecracker",
		"--api-sock", vm.APISock,
		"--enable-pci",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	vm.Cmd = cmd

	// Cleanup TAP & socket when process exits
	go func() {
		cmd.Wait()
		exec.Command("ip", "link", "del", vm.TapName).Run()
		os.Remove(vm.APISock)
	}()

	return nil
}

// StartVM configures VM via Firecracker API
func StartVM(vm *domain.VM, kernel, rootfs, inputDrive string) error {
	client := client.NewClient(vm.APISock)

	// Generate random MAC
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	r1 := rng.Intn(256)
	r2 := rng.Intn(256)
	r3 := rng.Intn(256)
	r4 := rng.Intn(256)
	mac := fmt.Sprintf("AA:FC:%02X:%02X:%02X:%02X", r1, r2, r3, r4)

	// Attach network
	if err := client.Put("/network-interfaces/eth0", []byte(fmt.Sprintf(`{
		"iface_id": "eth0",
		"host_dev_name": "%s",
		"guest_mac": "%s"
	}`, vm.TapName, mac))); err != nil {
		return fmt.Errorf("failed to attach network interface: %w", err)
	}

	// Machine config
	if err := client.Put("/machine-config", []byte(`{
		"vcpu_count": 1,
		"mem_size_mib": 512
	}`)); err != nil {
		return fmt.Errorf("failed to do machine config: %w", err)
	}

	// Boot source
	if err := client.Put("/boot-source", []byte(fmt.Sprintf(`{
		"kernel_image_path": "%s",
		"boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
	}`, kernel))); err != nil {
		return fmt.Errorf("failed to booot source: %w", err)
	}

	// Rootfs
	if err := client.Put("/drives/rootfs", []byte(fmt.Sprintf(`{
		"drive_id": "rootfs",
		"path_on_host": "%s",
		"is_root_device": true,
		"is_read_only": false
	}`, rootfs))); err != nil {
		return fmt.Errorf("failed attach rootfs : %w", err)
	}

	// Optional input drive (read-only)
	if inputDrive != "" {
		if err := client.Put("/drives/input_drive", []byte(fmt.Sprintf(`{
			"drive_id": "input_drive",
			"path_on_host": "%s",
			"is_root_device": false,
			"is_read_only": true
		}`, inputDrive))); err != nil {
			return fmt.Errorf("failed to attach input drive: %w", err)
		}
	}

	// Start instance
	if err := client.Put("/actions", []byte(`{
		"action_type": "InstanceStart"
	}`)); err != nil {
		return fmt.Errorf("failed to start instance: %w", err.(error))
	}
	return nil
}
