package instance

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/zanz1n/mc-manager/internal/distribution"
	"github.com/zanz1n/mc-manager/internal/dto"
	"github.com/zanz1n/mc-manager/internal/pb"
)

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
	TiB = GiB * 1024
)

type Event struct {
	Type pb.EventType `json:"type"`
	Data []byte       `json:"data"`
}

type InstanceLimits struct {
	// when it is <= 0 the default value will be used
	ShutdownAfterIdle time.Duration `json:"shutdown_after_idle"`
	// if the instance must be shutted down after some time idle
	AutoShutdown bool `json:"auto_shutdown"`

	MaxPlayers int32 `json:"max_players"`

	// 100 = 1 core
	CPU uint16 `json:"cpu" validate:"gte=0,lte=6400"`
	// in Bytes
	//
	// min: 512 MiB, max: 512 GiB
	RAM uint64 `json:"ram" validate:"gte=536870912,lte=549755813888"`
}

type InstanceConfig struct {
	Difficulty string `json:"difficulty"`
	Admin      string `json:"admin" validate:"required"`

	Port uint16 `json:"port" validate:"required"`

	ViewDistance       uint8 `json:"view_distance"`
	SimulationDistance uint8 `json:"simulation_distance"`

	AllowPirate bool `json:"allow_pirate"`
	PVP         bool `json:"pvp"`
}

type InstanceCreateData struct {
	ID   dto.Snowflake `json:"id" validate:"required"`
	Name string        `json:"name" validate:"required"`

	Version distribution.Version `json:"version"`
	Limits  InstanceLimits       `json:"limits"`
	Config  InstanceConfig       `json:"config"`
}

func newInstance(data InstanceCreateData) (*Instance, error) {
	err := validate.Struct(&data)
	if err != nil {
		return nil, errors.Join(ErrInvalidCreateData, err)
	}

	now := time.Now().Round(time.Millisecond)

	return &Instance{
		ID:         data.ID,
		LaunchedAt: now,
		Name:       data.Name,
		Version:    data.Version,
		Limits:     data.Limits,
		Config:     data.Config,
		listeners:  make(map[chan<- Event]struct{}),
	}, nil
}

type Instance struct {
	ID          dto.Snowflake
	ContainerID string
	LaunchedAt  time.Time
	Name        string

	Players  atomic.Int32
	Launched atomic.Bool

	Version distribution.Version
	Limits  InstanceLimits
	Config  InstanceConfig

	state atomic.Int32

	listeners map[chan<- Event]struct{}
	stream    types.HijackedResponse
	mu        sync.Mutex
}

func (i *Instance) SendCommand(cmd string) error {
	if len(cmd) == 0 {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()

	if cmd[len(cmd)-1] != '\n' {
		cmd += "\n"
	}

	_, err := i.stream.Conn.Write([]byte(cmd))
	if err != nil {
		err = errors.Join(ErrSendCommand, err)
	}
	return err
}

func (i *Instance) AttachListener(ch chan Event) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.listeners[ch] = struct{}{}
}

func (i *Instance) DetachListener(ch chan Event) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	_, ok := i.listeners[ch]
	delete(i.listeners, ch)
	return ok
}

func (i *Instance) GetState() pb.InstanceState {
	return pb.InstanceState(i.state.Load())
}

func (i *Instance) SetState(state pb.InstanceState) {
	i.state.Store(int32(state))
}

func (i *Instance) SendEvent(e Event) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for ch, _ := range i.listeners {
		counter := time.Tick(10 * time.Millisecond)
		select {
		case ch <- e:
		case <-counter:
		}
	}
}

func (i *Instance) launch() {
	i.Launched.Store(true)
	go i.backgroundLogs()
}

func (i *Instance) setStream(s types.HijackedResponse) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.stream = s
}

func (i *Instance) backgroundLogs() {
	for {
		line, _, err := i.stream.Reader.ReadLine()
		if err != nil {
			slog.Info("Instance: Logs closed", "id", i.ID, "error", err)
			break
		}

		i.SendEvent(Event{Type: pb.EventType_EVENT_LOG, Data: line})
	}
}

func (i *Instance) close() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for ch, _ := range i.listeners {
		counter := time.Tick(10 * time.Millisecond)
		select {
		case ch <- Event{Type: pb.EventType_EVENT_STOPPED}:
		case <-counter:
		}
		close(ch)
	}
}
