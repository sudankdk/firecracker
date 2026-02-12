package sandboxing

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/sudankdk/firecracker/internal/domain"
	client "github.com/sudankdk/firecracker/internal/httpClient"
)

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
func CreateTAP(tapName string) error {
	if err := exec.Command("ip", "tuntap", "add", tapName, "mode", "tap").Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "link", "set", tapName, "up").Run(); err != nil {
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

func configureVM(vm *domain.VM, kernel, rootfs, inputDrive string) error {
	httpClient := client.NewClient(vm.APISock)

	// Random MAC for eth0
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// mac := fmt.Sprintf("AA:FC:%02X:%02X:%02X:%02X", r.Intn(256), r.Intn(256), r.Intn(256), r.Intn(256))

	// // Network interface
	// if err := client.Put(httpClient, "/network-interfaces/eth0", []byte(fmt.Sprintf(`{
	// 	"iface_id": "eth0",
	// 	"host_dev_name": "%s",
	// 	"guest_mac": "%s"
	// }`, vm.TapName, mac))); err != nil {
	// 	return errors.New("failed to attach network interface")
	// }

	// Machine config
	if err := client.Put(httpClient, "/machine-config", []byte(`{
		"vcpu_count": 1,
		"mem_size_mib": 512
	}`)); err != nil {
		return errors.New("failed to configure machine")
	}

	// Boot source
	if err := client.Put(httpClient, "/boot-source", []byte(fmt.Sprintf(`{
		"kernel_image_path": "%s",
		"boot_args": "console=ttyS0 reboot=k panic=1 pci=off ip=off"
	}`, kernel))); err != nil {
		return errors.New("failed to configure boot source")
	}

	// Rootfs (read-only for isolation)
	if err := client.Put(httpClient, "/drives/rootfs", []byte(fmt.Sprintf(`{
		"drive_id": "rootfs",
		"path_on_host": "%s",
		"is_root_device": true,
		"is_read_only": true
	}`, rootfs))); err != nil {
		return errors.New("failed to configure rootfs")
	}

	// Input drive (uploaded file)
	if err := client.Put(httpClient, "/drives/input_drive", []byte(fmt.Sprintf(`{
		"drive_id": "input_drive",
		"path_on_host": "%s",
		"is_root_device": false,
		"is_read_only": true
	}`, inputDrive))); err != nil {
		return errors.New("failed to configure input drive")
	}

	// Start instance
	if err := client.Put(httpClient, "/actions", []byte(`{"action_type":"InstanceStart"}`)); err != nil {
		return errors.New("failed to start instance")
	}
	return nil
}
