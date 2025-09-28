package kv

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/zanz1n/mc-manager/internal/utils"
)

var _ KVStorer = (*LocalKV)(nil)

type LocalValue struct {
	Object     any
	Expiration int64
}

func newLocalValue(obj any, exp time.Duration) LocalValue {
	return LocalValue{
		Object:     obj,
		Expiration: time.Now().Add(exp).UnixMilli(),
	}
}

func (v *LocalValue) assignTo(to any) bool {
	toValue := reflect.ValueOf(to)
	if toValue.Kind() != reflect.Pointer {
		return false
	}

	toElem := toValue.Elem()
	if !toElem.CanSet() {
		return false
	}

	fromValue := reflect.ValueOf(v.Object)
	if fromValue.Kind() == reflect.Pointer {
		fromValue = fromValue.Elem()
	}

	if !toElem.Type().AssignableTo(fromValue.Type()) {
		return false
	}

	toElem.Set(fromValue)
	return true
}

func (v *LocalValue) String() string {
	if str, ok := v.Object.(string); ok {
		return str
	}

	b, err := json.Marshal(v)
	if err != nil {
		return "<nil>"
	}
	return utils.UnsafeString(b)
}

func (v *LocalValue) isExpired() bool {
	if v.Expiration <= 0 {
		return false
	}

	return time.Now().UnixMilli() > v.Expiration
}

type LocalKV struct {
	m  map[string]LocalValue
	mu sync.RWMutex

	fileName string
	closeIC  chan struct{}
	closeFB  chan struct{}
}

func NewLocalKV(fileName string, saveInterval time.Duration) *LocalKV {
	kv := &LocalKV{
		m:        make(map[string]LocalValue),
		mu:       sync.RWMutex{},
		fileName: fileName,
		closeIC:  make(chan struct{}),
		closeFB:  make(chan struct{}),
	}

	if fileName != "" {
		err := kv.LoadFrom(fileName)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Error(
				"LocalKV: Failed to load data",
				"file", fileName,
				"error", err,
			)
		}

		if saveInterval > 0 {
			go kv.launchBackground(saveInterval)
		}
	}

	return kv
}

func (l *LocalKV) launchBackground(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			if err := l.DumpTo(l.fileName); err != nil {
				slog.Warn(
					"LocalKV: Failed to dump",
					"file", l.fileName,
					"took", time.Since(now).Round(time.Microsecond),
					"error", err,
				)
			} else {
				slog.Debug(
					"LocalKV: Dumped",
					"file", l.fileName,
					"took", time.Since(now).Round(time.Microsecond),
				)
			}

		case <-l.closeIC:
			l.closeFB <- struct{}{}
			return
		}
	}
}

func (l *LocalKV) DumpTo(fpath string) error {
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UnixMilli()
	for k, v := range l.m {
		if now > v.Expiration {
			delete(l.m, k)
		}
	}

	if err = gob.NewEncoder(file).Encode(l.m); err != nil {
		return err
	}
	return nil
}

func (l *LocalKV) LoadFrom(fpath string) error {
	file, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	l.mu.Lock()
	defer l.mu.Unlock()

	if err = gob.NewDecoder(file).Decode(&l.m); err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	for k, v := range l.m {
		if now > v.Expiration {
			delete(l.m, k)
		}
	}
	return nil
}

func (l *LocalKV) get(key string) (LocalValue, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	v, ok := l.m[key]
	if !ok {
		return LocalValue{}, ErrValueNotFound
	}

	if v.isExpired() {
		return LocalValue{}, ErrValueNotFound
	}

	return v, nil
}

func (l *LocalKV) getEx(key string, ttl time.Duration) (LocalValue, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.m[key]
	if !ok {
		return LocalValue{}, ErrValueNotFound
	}

	if v.isExpired() {
		return LocalValue{}, ErrValueNotFound
	}

	newV := newLocalValue(v.Object, ttl)
	l.m[key] = newV
	return newV, nil
}

func (l *LocalKV) set(key string, value LocalValue) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.m[key] = value
}

func (l *LocalKV) del(key string) bool {
	l.mu.RLock()
	_, ok := l.m[key]
	l.mu.RUnlock()
	if !ok {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.m, key)

	return true
}

// Exists implements KVStorer.
func (l *LocalKV) Exists(ctx context.Context, key string) (bool, error) {
	_, err := l.get(key)
	return err == nil, nil
}

// Get implements KVStorer.
func (l *LocalKV) Get(ctx context.Context, key string) (string, error) {
	v, err := l.get(key)
	if err != nil {
		return "", err
	}

	return v.String(), nil
}

// GetEx implements KVStorer.
func (l *LocalKV) GetEx(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (string, error) {
	v, err := l.get(key)
	if err != nil {
		return "", err
	}

	return v.String(), nil
}

// GetValue implements KVStorer.
func (l *LocalKV) GetValue(ctx context.Context, key string, v any) error {
	localV, err := l.get(key)
	if err != nil {
		return err
	}

	if ok := localV.assignTo(v); !ok {
		return ErrValueNotPointer
	}
	return nil
}

// GetValueEx implements KVStorer.
func (l *LocalKV) GetValueEx(
	ctx context.Context,
	key string,
	ttl time.Duration,
	v any,
) error {
	localV, err := l.getEx(key, ttl)
	if err != nil {
		return err
	}

	if ok := localV.assignTo(v); !ok {
		return ErrValueNotPointer
	}
	return nil
}

// Set implements KVStorer.
func (l *LocalKV) Set(ctx context.Context, key string, value string) error {
	l.set(key, newLocalValue(value, -1))
	return nil
}

// SetEx implements KVStorer.
func (l *LocalKV) SetEx(
	ctx context.Context,
	key string,
	value string,
	ttl time.Duration,
) error {
	l.set(key, newLocalValue(value, ttl))
	return nil
}

// SetValue implements KVStorer.
func (l *LocalKV) SetValue(ctx context.Context, key string, v any) error {
	l.set(key, newLocalValue(v, -1))
	return nil
}

// SetValueEx implements KVStorer.
func (l *LocalKV) SetValueEx(
	ctx context.Context,
	key string,
	v any,
	ttl time.Duration,
) error {
	l.set(key, newLocalValue(v, ttl))
	return nil
}

// Delete implements KVStorer.
func (l *LocalKV) Delete(ctx context.Context, key string) error {
	if ok := l.del(key); !ok {
		return ErrValueNotFound
	}
	return nil
}

// Close implements KVStorer.
func (l *LocalKV) Close() error {
	if l.fileName != "" {
		l.closeIC <- struct{}{}
		<-l.closeFB

		return l.DumpTo(l.fileName)
	}
	return nil
}
