package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"github.com/olahol/melody"
)

var apiVersion = "b537c464"
var containerPortMap = make(map[string]int)
var portMutex sync.Mutex
var currentPort = 8001

//go:embed index.html node_modules/xterm/css/xterm.css node_modules/xterm/lib/xterm.js
var content embed.FS

// execCmd executes a system command and returns its output as a string.
func execCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// executeContainerAction executes a specified LXC command and sends a JSON response.
func executeContainerAction(w http.ResponseWriter, command string, successMessage string, args ...string) {
	_, err := execCmd(command, args...)
	if err != nil {
		sendJSONResponse(w, &Response{StatusCode: http.StatusInternalServerError, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, &Response{StatusCode: http.StatusOK, Message: successMessage}, http.StatusOK)
}

// executeContainerActionWithArgs is similar to executeContainerAction but allows passing a slice of arguments.
func executeContainerActionWithArgs(w http.ResponseWriter, command string, args []string, successMessage string) {
	_, err := execCmd(command, args...)
	if err != nil {
		sendJSONResponse(w, &Response{StatusCode: http.StatusInternalServerError, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, &Response{StatusCode: http.StatusOK, Message: successMessage}, http.StatusOK)
}

// sendJSONResponse is a helper function to marshal data to JSON and write it to the response writer.
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
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

// GetVersion handles the HTTP request to get the LXC version.
func GetVersion(w http.ResponseWriter, r *http.Request) {
	version, err := execCmd("lxc-info", "--version")
	if err != nil {
		http.Error(w, "Failed to get LXC version", http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, &VersionResponse{Version: version}, http.StatusOK)
}

// GetContainers handles the HTTP request to list LXC containers.
func GetContainers(w http.ResponseWriter, r *http.Request) {
	output, err := execCmd("lxc-ls")
	if err != nil {
		http.Error(w, "Failed to list containers", http.StatusInternalServerError)
		return
	}
	containers := strings.Split(output, "\n")
	sendJSONResponse(w, &ContainersResponse{Containers: containers}, http.StatusOK)
}

// GetContainerInfo handles the HTTP request to get information about a specific container.
func GetContainerInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	output, err := execCmd("lxc-info", "--name", containerName)
	if err != nil {
		http.Error(w, "Failed to get container info", http.StatusInternalServerError)
		return
	}
	info := parseContainerInfo(output)
	sendJSONResponse(w, &ContainerResponse{Containers: []ContainerInfo{info}}, http.StatusOK)
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

// StartContainer handles the HTTP request to start a specific LXC container.
func StartContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	executeContainerActionWithArgs(w, "lxc-start", []string{containerName}, "Container started successfully")
}

// StopContainer handles the HTTP request to stop a specific LXC container.
func StopContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	executeContainerActionWithArgs(w, "lxc-stop", []string{containerName}, "Container stopped successfully")
}

// FreezeContainer handles the HTTP request to freeze a specific LXC container.
func FreezeContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	executeContainerActionWithArgs(w, "lxc-freeze", []string{containerName}, "Container frozen successfully")
}

// UnfreezeContainer handles the HTTP request to unfreeze a specific LXC container.
func UnfreezeContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	executeContainerActionWithArgs(w, "lxc-unfreeze", []string{containerName}, "Container unfrozen successfully")
}

// CreateContainer handles the HTTP request to create a new LXC container.
func CreateContainer(w http.ResponseWriter, r *http.Request) {
	// Assume ContainerCreateRequest struct has been defined according to your requirements.
	var request ContainerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONResponse(w, &Response{StatusCode: http.StatusBadRequest, Message: "Invalid request format"}, http.StatusBadRequest)
		return
	}

	args := []string{"-t", request.Template, "-n", request.ContainerName, "--", "--server", request.ImageSource, "--dist", request.Distribution, "--release", request.Release, "--arch", request.Architecture}
	executeContainerActionWithArgs(w, "lxc-create", args, "Container created successfully")
}

// DestroyContainer handles the HTTP request to destroy an existing LXC container.
func DestroyContainer(w http.ResponseWriter, r *http.Request) {
	var request ContainerDestroyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONResponse(w, &Response{StatusCode: http.StatusBadRequest, Message: "Invalid request format"}, http.StatusBadRequest)
		return
	}

	executeContainerAction(w, "lxc-destroy", "Container destroyed successfully", "-n", request.ContainerName)
}

// The port pool is used to keep track of the ports that are in use.
func getAvailablePort() int {
	portMutex.Lock()
	defer portMutex.Unlock()

	if currentPort > 8999 {
		currentPort = 8001
	}

	port := currentPort
	currentPort++
	return port
}

// releasePort releases a port from the port pool.
func releasePort(port int) {
	_, err := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port)).Output()
	if err != nil {
		fmt.Printf("Error running lsof command: %v\n", err)
		return
	}

	pid, err := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port)).Output()
	if err != nil {
		fmt.Printf("Error running lsof command: %v\n", err)
		return
	}

	if pid == nil {
		fmt.Printf("No process found listening on port %d\n", port)
		return
	}

	_, err = exec.Command("kill", "-9", string(pid)).Output()
	if err != nil {
		fmt.Printf("Error killing process on port %d: %v\n", port, err)
		return
	}

	fmt.Printf("Successfully killed process listening on port %d\n", port)

	for name, p := range containerPortMap {
		if p == port {
			delete(containerPortMap, name)
			break
		}
	}
}

// isAttach checks whether the container is attached to a websocket.
func isAttach(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	_, ok := containerPortMap[containerName]
	if ok {
		sendJSONResponse(w, &IsAttach{IsAttach: true}, http.StatusOK)
		return
	}
	sendJSONResponse(w, &IsAttach{IsAttach: false}, http.StatusOK)
}

// AttachContainer handles the HTTP request to attach to a specific LXC container.
func AttachContainer(w http.ResponseWriter, r *http.Request, m *melody.Melody) {
	vars := mux.Vars(r)
	containerName := vars["name"]
	port := getAvailablePort()

	if port == -1 {
		http.Error(w, "No available ports in the pool", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("lxc-attach", containerName, "/bin/login")
	f, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		f.Write(msg)
	})

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

	fs := http.FileServer(http.FS(content))
	http.Handle("/", http.StripPrefix("/", fs))

	go func() {
		err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), nil)
		if err != nil {
			releasePort(port)
			return
		}
	}()

	containerPortMap[containerName] = port

	attachInfo := AttachInfo{
		ContainerName: containerName,
		Port:          port,
	}

	sendJSONResponse(w, attachInfo, http.StatusOK)
}

// DetachContainer handles the HTTP request to detach from a specific LXC container.
func DetachContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerName := vars["name"]

	port, ok := containerPortMap[containerName]
	if !ok {
		http.Error(w, "Container not found in the port map", http.StatusNotFound)
		return
	}

	delete(containerPortMap, containerName)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: addr}
	go func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("Error closing server on port %d: %v\n", port, err)
		}
	}()

	releasePort(port)

	w.WriteHeader(http.StatusOK)
}

func createServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}

func main() {
	r := mux.NewRouter()
	m := melody.New()
	r.HandleFunc("/webterminal", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	}).Methods("GET")

	r.HandleFunc("/version", GetVersion).Methods("GET")
	r.HandleFunc("/apiversion", GetAPIVersion).Methods("GET")
	r.HandleFunc("/containers", GetContainers).Methods("GET")

	containerRouter := r.PathPrefix("/container/{name}").Subrouter()
	containerRouter.HandleFunc("", GetContainerInfo).Methods("GET")
	containerRouter.HandleFunc("/start", StartContainer).Methods("POST")
	containerRouter.HandleFunc("/stop", StopContainer).Methods("POST")
	containerRouter.HandleFunc("/freeze", FreezeContainer).Methods("POST")
	containerRouter.HandleFunc("/unfreeze", UnfreezeContainer).Methods("POST")
	containerRouter.HandleFunc("/isattach", isAttach).Methods("GET")
	containerRouter.HandleFunc("/attach", func(w http.ResponseWriter, r *http.Request) {
		AttachContainer(w, r, m)
	}).Methods("GET")
	containerRouter.HandleFunc("/detach", DetachContainer).Methods("GET")

	r.HandleFunc("/add/container", CreateContainer).Methods("POST")
	r.HandleFunc("/del/container", DestroyContainer).Methods("POST")

	caCertFile := "/data/lxc/lxc-api-ca.crt"
	serverCertFile := "/data/lxc/lxc-api.crt"
	serverKeyFile := "/data/lxc/lxc-api.key"

	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		log.Fatal("ERROR: Failed to read CA certificate file", err)
	}

	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		log.Fatal("ERROR: Failed to load server certificate and key", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	srv := createServer("0.0.0.0:8000", r)
	srv.TLSConfig = tlsConfig

	log.Fatal(srv.ListenAndServeTLS(serverCertFile, serverKeyFile))

}
