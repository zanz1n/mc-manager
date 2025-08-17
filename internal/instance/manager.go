package instance

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/zanz1n/mc-manager/internal/dto"
)

type Manager struct {
	m  map[dto.Snowflake]*Instance
	mu sync.RWMutex

	runner Runner
}

func NewManager(runner Runner) *Manager {
	return &Manager{
		m:      make(map[dto.Snowflake]*Instance),
		runner: runner,
	}
}

func (m *Manager) Launch(ctx context.Context, data InstanceCreateData) (*Instance, error) {
	start := time.Now()

	if _, err := m.GetById(ctx, data.ID); err == nil {
		return nil, errors.Join(
			ErrInstanceAlreadyLaunched,
			errors.New(data.ID.String()),
		)
	}

	i, err := newInstance(data)
	if err != nil {
		return nil, err
	}
	m.insert(i)

	err = m.runner.Create(ctx, i)
	if err != nil {
		m.remove(i.ID)
		slog.Error(
			"Manager: Failed to create instance",
			"took", time.Since(start).Round(time.Microsecond),
			"error", err,
		)
		return nil, err
	}

	err = m.runner.Launch(ctx, i)
	if err != nil {
		m.remove(i.ID)
		slog.Error(
			"Manager: Failed to launch instance",
			"id", i.ID,
			"took", time.Since(start).Round(time.Microsecond),
			"error", err,
		)
		return nil, err
	}

	slog.Info(
		"Manager: Launched instance",
		"id", i.ID,
		"took", time.Since(start).Round(time.Microsecond),
	)

	return i, nil
}

func (m *Manager) GetById(ctx context.Context, id dto.Snowflake) (*Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.m[id]
	if !ok {
		return nil, errors.Join(
			ErrInstanceNotFound,
			errors.New(id.String()),
		)
	}

	return instance, nil
}

func (m *Manager) GetMany(ctx context.Context, ids []uint64) ([]*Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*Instance, len(ids))
	for j, idb := range ids {
		id := dto.Snowflake(idb)
		i, ok := m.m[id]
		if !ok {
			return nil, errors.Join(
				ErrInstanceNotFound,
				errors.New(id.String()),
			)
		}
		instances[j] = i
	}

	return instances, nil
}

func (m *Manager) Stop(ctx context.Context, id dto.Snowflake) error {
	start := time.Now()

	i, err := m.GetById(ctx, id)
	if err != nil {
		return err
	}

	err = m.runner.Stop(ctx, i)
	if err != nil {
		slog.Error(
			"Manager: Failed to stop instance",
			"id", id,
			"took", time.Since(start).Round(time.Microsecond),
			"error", err,
		)
	} else {
		slog.Info(
			"Manager: Stopped instance",
			"id", id,
			"took", time.Since(start).Round(time.Microsecond),
		)
	}

	m.remove(id)
	return err
}

func (m *Manager) insert(i *Instance) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.m[i.ID] = i
}

func (m *Manager) remove(id dto.Snowflake) (*Instance, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	i, ok := m.m[id]
	if ok {
		delete(m.m, id)
	}

	return i, ok
}
