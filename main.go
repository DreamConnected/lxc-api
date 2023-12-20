package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

//go:embed index.html node_modules/xterm/css/xterm.css node_modules/xterm/lib/xterm.js
var content embed.FS
var apiVersion = "b537c464"

type VersionResponse struct {
	Version string `json:"version"`
}

func GetAPIVersion(w http.ResponseWriter, r *http.Request) {
	resp := &VersionResponse{
		Version: apiVersion,
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

type ContainerResponse struct {
	Containers []ContainerInfo `json:"containers"`
}

type ContainersResponse struct {
	Containers []string `json:"containers"`
}

type ContainerInfo struct {
	Name       string         `json:"name"`
	State      string         `json:"state"`
	PID        string         `json:"pid"`
	IP         string         `json:"ip"`
	CPUUsage   string         `json:"cpu_usage"`
	BlkIOUsage string         `json:"blkio_usage"`
	MemoryUse  string         `json:"memory_use"`
	KMemUse    string         `json:"kmem_use"`
	Link       string         `json:"link"`
	LinkState  LinkStatistics `json:"link_state"`
}

type LinkStatistics struct {
	TXBytes    string `json:"tx_bytes"`
	RXBytes    string `json:"rx_bytes"`
	TotalBytes string `json:"total_bytes"`
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type ContainerCreateRequest struct {
	Template      string `json:"template"`
	ContainerName string `json:"container_name"`
	ImageSource   string `json:"image_source"`
	Distribution  string `json:"distribution"`
	Release       string `json:"release"`
	Architecture  string `json:"architecture"`
}

type ContainerDestroyRequest struct {
	ContainerName string `json:"del_container"`
}

func GetVersion(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("lxc-info", "--version")

	output, err := cmd.Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	version := strings.TrimSuffix(string(output), "\n")

	resp := &VersionResponse{
		Version: version,
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func GetContainers(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("lxc-ls")

	output, err := cmd.Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	containers := strings.Split(strings.TrimSpace(string(output)), "\n")
	resp := &ContainersResponse{
		Containers: containers,
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func GetContainerInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]

	cmd := exec.Command("lxc-info", "--name", containerName)

	output, err := cmd.Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info := parseContainerInfo(string(output))

	resp := &ContainerResponse{
		Containers: []ContainerInfo{info},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func parseContainerInfo(output string) ContainerInfo {
	lines := strings.Split(output, "\n")
	info := ContainerInfo{}
	link := LinkStatistics{}

	for i := 0; i < len(lines); i++ {
		fields := strings.SplitN(lines[i], ":", 2)
		if len(fields) != 2 {
			continue
		}

		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "Name":
			info.Name = value
		case "State":
			info.State = value
		case "PID":
			info.PID = value
		case "IP":
			info.IP = value
		case "CPU use":
			info.CPUUsage = value
		case "BlkIO use":
			info.BlkIOUsage = value
		case "Memory use":
			info.MemoryUse = value
		case "KMem use":
			info.KMemUse = value
		case "Link":
			info.Link = value
		case "TX bytes":
			link.TXBytes = value
		case "RX bytes":
			link.RXBytes = value
		case "Total bytes":
			link.TotalBytes = value
		}
	}

	info.LinkState = link

	return info
}

func StartContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]

	cmd := exec.Command("lxc-start", containerName)

	err := cmd.Run()
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}

		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Message:    "Container started successfully",
	}

	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func StopContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]

	cmd := exec.Command("lxc-stop", containerName)

	err := cmd.Run()
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}

		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Message:    "Container stopped successfully",
	}

	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func FreezeContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]

	cmd := exec.Command("lxc-freeze", containerName)

	err := cmd.Run()
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}

		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Message:    "Container frozen successfully",
	}

	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func UnfreezeContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]

	cmd := exec.Command("lxc-unfreeze", containerName)

	err := cmd.Run()
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}

		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Message:    "Container unfrozen successfully",
	}

	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func CreateContainer(w http.ResponseWriter, r *http.Request) {
	var request ContainerCreateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request format",
		}
		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	cmd := exec.Command("lxc-create",
		"-t", request.Template,
		"-n", request.ContainerName,
		"--",
		"--server", request.ImageSource,
		"--dist", request.Distribution,
		"--release", request.Release,
		"--arch", request.Architecture,
	)

	err = cmd.Run()
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Message:    "Container created successfully",
	}
	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func DestroyContainer(w http.ResponseWriter, r *http.Request) {
	var request ContainerDestroyRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request format",
		}
		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	cmd := exec.Command("lxc-destroy", "-n", request.ContainerName)

	err = cmd.Run()
	if err != nil {
		resp := &Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
		jsonData, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Message:    "Container destroyed successfully",
	}
	jsonData, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

var templates *template.Template

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Println("Received message:", string(message))

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/version", GetVersion).Methods("GET")

	r.HandleFunc("/apiversion", GetAPIVersion).Methods("GET")

	r.HandleFunc("/containers", GetContainers).Methods("GET")

	r.HandleFunc("/container/{name}", GetContainerInfo).Methods("GET")

	r.HandleFunc("/container/{name}/start", StartContainer).Methods("POST")

	r.HandleFunc("/container/{name}/stop", StopContainer).Methods("POST")

	r.HandleFunc("/container/{name}/freeze", FreezeContainer).Methods("POST")

	r.HandleFunc("/container/{name}/unfreeze", UnfreezeContainer).Methods("POST")

	r.HandleFunc("/add/container", CreateContainer).Methods("POST")

	r.HandleFunc("/del/container", DestroyContainer).Methods("POST")

	c := exec.Command("sh")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	m := melody.New()

	go func() {
		for {
			buf := make([]byte, 1024)
			read, err := f.Read(buf)
			if err != nil {
				return
			}
			m.Broadcast(buf[:read])
		}
	}()

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		f.Write(msg)
	})

	http.HandleFunc("/webterminal", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	fs := http.FileServer(http.FS(content))
	http.Handle("/", http.StripPrefix("/", fs))

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Fatal(srv.ListenAndServeTLS("/data/lxc/lxc-api.crt", "/data/lxc/lxc-api.key"))
	}()

	log.Fatal(http.ListenAndServe("0.0.0.0:8001", nil))
}
