#!/bin/bash

# Clean up any existing Firecracker processes and sockets
sudo pkill -9 firecracker 2>/dev/null || true
sudo rm -f /tmp/firecracker.sock

# Start Firecracker in the background
sudo ./firecracker --api-sock /tmp/firecracker.sock &

# Wait for socket to be created
for i in {1..10}; do
    if [ -S /tmp/firecracker.sock ]; then
        echo "Socket created, fixing permissions..."
        sudo chmod 666 /tmp/firecracker.sock
        echo "Firecracker is ready! You can now run your Go program."
        exit 0
    fi
    sleep 0.5
done

echo "Error: Socket not created in time"
exit 1
