package controlapi

const (
	APIPrefix = "$NEX"
)

const (
	InfoResponseType = "io.nats.nex.v1.info_response"
	PingResponseType = "io.nats.nex.v1.ping_response"
	RunResponseType  = "io.nats.nex.v1.run_response"
	TagOS            = "nex.os"
	TagArch          = "nex.arch"
	TagCPUs          = "nex.cpucount"
)

type RunResponse struct {
	Started   bool   `json:"started"`
	MachineId string `json:"machine_id"`
	PublicKey string `json:"public_key"`
	Issuer    string `json:"issuer"`
	Hash      string `json:"hash"`
}

type PingResponse struct {
	NodeId          string            `json:"node_id"`
	Version         string            `json:"version"`
	Uptime          string            `json:"uptime"`
	RunningMachines int               `json:"running_machines"`
	Tags            map[string]string `json:"tags,omitempty"`
}

type InfoResponse struct {
	Version    string            `json:"version"`
	Uptime     string            `json:"uptime"`
	PublicXKey string            `json:"public_xkey"`
	Tags       map[string]string `json:"tags,omitempty"`
	Machines   []MachineSummary  `json:"machines"`
}

type MachineSummary struct {
	Id       string          `json:"id"`
	Healthy  bool            `json:"healthy"`
	Uptime   string          `json:"uptime"`
	Workload WorkloadSummary `json:"workload,omitempty"`
}

type WorkloadSummary struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Runtime      string `json:"runtime"`
	WorkloadType string `json:"type"`
	Hash         string `json:"hash"`
}

type Envelope struct {
	PayloadType string      `json:"type"`
	Data        interface{} `json:"data,omitempty"`
	Error       interface{} `json:"error,omitempty"`
}

func NewEnvelope(dataType string, data interface{}, err *string) Envelope {
	var e interface{}
	if err != nil {
		e = *err
	}
	return Envelope{
		PayloadType: dataType,
		Data:        data,
		Error:       e,
	}
}