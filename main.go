package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/vm/start", func(w http.ResponseWriter, r *http.Request) {
		err := StartVM(
			"/tmp/firecracker.sock",
			"/mnt/d/firecracker/hello-vmlinux.bin",
			"/mnt/d/firecracker/hello-rootfs.ext4",
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("VM started"))
	})

	port := "8080"
	log.Printf("HTTP server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
