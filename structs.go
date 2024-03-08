package main

// API
type VersionResponse struct {
	Version string `json:"version"`
}

// Get Container Info
type ContainerResponse struct {
	Containers []ContainerInfo `json:"containers"`
}

// Get Container Lists
type ContainersResponse struct {
	Containers []string `json:"containers"`
}

// Get Container Field
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

// Get Container Field of Network
type LinkStatistics struct {
	TXBytes    string `json:"tx_bytes"`
	RXBytes    string `json:"rx_bytes"`
	TotalBytes string `json:"total_bytes"`
}

// HTTP Response
type Response struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// Request for Container Creation
type ContainerCreateRequest struct {
	Template      string `json:"template"`
	ContainerName string `json:"container_name"`
	ImageSource   string `json:"image_source"`
	Distribution  string `json:"distribution"`
	Release       string `json:"release"`
	Architecture  string `json:"architecture"`
}

// Request for container Destroy
type ContainerDestroyRequest struct {
	ContainerName string `json:"del_container"`
}

type AttachInfo struct {
	ContainerName string `json:"containerName"`
	Port          int    `json:"port"`
}

type IsAttach struct {
	IsAttach bool `json:"isAttach"`
}
