package sandboxing

import (
	"fmt"
	"log"

	client "github.com/sudankdk/firecracker/internal/httpClient"
)

// StopVM sends shutdown signal to Firecracker VM
func StopVM(jobID string) error {
	sock := fmt.Sprintf("/tmp/firecracker-%s.sock", jobID)
	httpClient := client.NewClient(sock)

	// Send InstanceShutdown action
	if err := client.Put(httpClient, "/actions", []byte(`{
		"action_type": "SendCtrlAltDel"
	}`)); err != nil {
		return fmt.Errorf("failed to send shutdown signal: %w", err)
	}

	log.Printf("Shutdown signal sent to VM for job %s", jobID)
	return nil
}
