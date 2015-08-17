package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"io/ioutil"
	"os"
	"io"
)

type UploadStatus struct {
	Path	  string	`json:"path"`
}
type ConfigFile struct {
	AllowedIps []string
}

func contains(s []string, e string) bool {
	for _, a := range s { if a == e { return true } }
	return false
}

const FilesFolderPath string = `./files/`
const ConfigFilePath string = `./config.json`
const DefaultConfig string = `{"AllowedIps": ["127.0.0.1", "::1"]}`
func isAllowed(r *http.Request) bool {
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		fmt.Printf("Config file does not exist. Creating default file.\n")
		ioutil.WriteFile(ConfigFilePath, []byte(DefaultConfig), 0644)
	}

	file, e := ioutil.ReadFile(ConfigFilePath)
	if e != nil {
		fmt.Printf("Config file error: %v\n", e)
		return false
	}

	var cfg ConfigFile
	json.Unmarshal(file, &cfg)

	ip, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		fmt.Printf("Failed to split ip: %v\n", r.RemoteAddr)
		return false
	}

	if contains(cfg.AllowedIps, ip) {
		return true
	}
	fmt.Printf("Warning: %v blocked from /upload\n", r.RemoteAddr)

	return false
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed(r) {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	f, err := os.OpenFile(FilesFolderPath + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	status := UploadStatus{}
	status.Path = "f/" + handler.Filename

	json.NewEncoder(w).Encode(status)
}

type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

type neuteredReaddirFile struct {
	http.File
}

func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func main() {
	if _, err := os.Stat(FilesFolderPath); os.IsNotExist(err) {
		fmt.Printf("Files folder does not exist. Creating.\n")
		os.Mkdir(FilesFolderPath, 0644)
	}

	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/f/", http.StripPrefix("/f/", http.FileServer(justFilesFilesystem{http.Dir(FilesFolderPath)})))
	http.ListenAndServe(":8080", nil)
}
