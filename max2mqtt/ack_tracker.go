package main

import (
	"log/slog"
	"sync"
	"time"
)

const configAckTimeout = 30 * time.Second

// PendingConfigAck tracks an outstanding config TX until the device Acks (CUL push).
type PendingConfigAck struct {
	DeviceID string
	MsgType  byte
	Field    string
	Previous DeviceConfigSnapshot
	Next     DeviceConfigSnapshot
	Timer    *time.Timer
}

// DeviceConfigSnapshot is a copy of config used for optimistic update / rollback.
type DeviceConfigSnapshot struct {
	Temps   DeviceConfig
	Valve   ValveConfig
	Display string
	GroupID int
}

type AckTracker struct {
	mutex   sync.Mutex
	pending map[string]*PendingConfigAck // key: deviceID
}

var ackTracker *AckTracker

func initAckTracker() {
	ackTracker = &AckTracker{
		pending: make(map[string]*PendingConfigAck),
	}
	slog.Info("AckTracker initialized", "timeout", configAckTimeout)
}

// TrackConfig registers a pending config change. Optimistically applies Next and publishes.
// On Ack from device, confirms; on timeout, restores Previous.
func (a *AckTracker) TrackConfig(deviceID string, msgType byte, field string, prev, next DeviceConfigSnapshot) {
	if a == nil {
		return
	}
	a.mutex.Lock()
	if old, ok := a.pending[deviceID]; ok && old.Timer != nil {
		old.Timer.Stop()
	}
	entry := &PendingConfigAck{
		DeviceID: deviceID,
		MsgType:  msgType,
		Field:    field,
		Previous: prev,
		Next:     next,
	}
	entry.Timer = time.AfterFunc(configAckTimeout, func() {
		a.handleTimeout(deviceID)
	})
	a.pending[deviceID] = entry
	a.mutex.Unlock()

	applyConfigSnapshot(deviceID, next)
	publishDeviceConfig(deviceID)
	slog.Info("Config change pending Ack", "device", deviceID, "field", field, "type", msgType)
}

// Confirm marks a successful device Ack/status push and cancels the timeout.
func (a *AckTracker) Confirm(deviceID string) {
	if a == nil {
		return
	}
	a.mutex.Lock()
	entry, ok := a.pending[deviceID]
	if !ok {
		a.mutex.Unlock()
		return
	}
	if entry.Timer != nil {
		entry.Timer.Stop()
	}
	delete(a.pending, deviceID)
	field := entry.Field
	a.mutex.Unlock()

	// Authoritative publish after Ack (cache already has Next from optimistic apply)
	publishDeviceConfig(deviceID)
	slog.Info("Config Ack confirmed", "device", deviceID, "field", field)
}

func (a *AckTracker) handleTimeout(deviceID string) {
	a.mutex.Lock()
	entry, ok := a.pending[deviceID]
	if !ok {
		a.mutex.Unlock()
		return
	}
	prev := entry.Previous
	field := entry.Field
	delete(a.pending, deviceID)
	a.mutex.Unlock()

	slog.Warn("Config Ack timeout; restoring previous values", "device", deviceID, "field", field)
	applyConfigSnapshot(deviceID, prev)
	publishDeviceConfig(deviceID)
}

func applyConfigSnapshot(deviceID string, snap DeviceConfigSnapshot) {
	cfg := getDeviceConfig(deviceID)
	valve := getValveConfig(deviceID)
	configMutex.Lock()
	*cfg = snap.Temps
	*valve = snap.Valve
	cfg.DisplayMode = snap.Display
	cfg.GroupID = snap.GroupID
	configMutex.Unlock()
}

func snapshotDevice(deviceID string) DeviceConfigSnapshot {
	cfg := getDeviceConfig(deviceID)
	valve := getValveConfig(deviceID)
	configMutex.RLock()
	defer configMutex.RUnlock()
	return DeviceConfigSnapshot{
		Temps:   *cfg,
		Valve:   *valve,
		Display: cfg.DisplayMode,
		GroupID: cfg.GroupID,
	}
}
