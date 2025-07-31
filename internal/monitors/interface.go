package monitors

import (
	"context"
	"time"
)

type MonitorType string

const (
	TypeREST    MonitorType = "rest"
	TypeGRPC    MonitorType = "grpc"
	TypeQuality MonitorType = "quality"
)

type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
	StatusInfo Status = "info"
)

type Monitor interface {
	Name() string
	Type() MonitorType
	Check(ctx context.Context) (*Result, error)
}

type Result struct {
	Name      string                 `json:"name"`
	Type      MonitorType            `json:"type"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}
