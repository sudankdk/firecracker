package main

import (
	"log"
	"net/http"
	"os"
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

	http.HandleFunc("uploadfile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get file from form", 500)
			return
		}
		defer file.Close()
		// Process the uploaded file (e.g., save it, hash it, etc.)
		os.WriteFile("uploaded_file", []byte("file content"), 0644)
		hash, err := hashFile("uploaded_file")
		if err != nil {
			http.Error(w, "Failed to hash file", 500)
			return
		}
		w.Write([]byte("File uploaded and hashed: " + hash))

	})

	port := "8080"
	log.Printf("HTTP server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
