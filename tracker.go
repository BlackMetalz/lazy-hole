package main

import (
	"sync"

	"golang.org/x/crypto/ssh"
)

// Global tracker instance
var effectTracker = NewEffectTracker() // Global variable, accessible from anywhere

// Effect types
const (
	EffectBlackHole  = "blackhole"
	EffectLatency    = "latency"
	EffectPacketLoss = "packetloss"
	// EffectPartition  = "partition"
	EffectPortBlock = "portblock"
)

// ActiveEffect represents one active effect on a host
type ActiveEffect struct {
	Type   string // blackhole, latency, packetloss, partition
	Target string // IP, CIDR, interface, source IP
	Value  string // delay, percentage
}

// EffectTracker tracks active effects per host
type EffectTracker struct {
	mu      sync.Mutex
	effects map[string][]ActiveEffect // hostname --> effects
}

// NewEffectTracker creates a new tracker
func NewEffectTracker() *EffectTracker {
	return &EffectTracker{
		effects: make(map[string][]ActiveEffect),
	}
}

// Add effect for a host
func (t *EffectTracker) Add(hostname string, effect ActiveEffect) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.effects[hostname] = append(t.effects[hostname], effect)
}

// Remove specific effect from host
func (t *EffectTracker) Remove(hostname string, effect ActiveEffect) {
	t.mu.Lock()
	defer t.mu.Unlock()

	effects := t.effects[hostname]
	for i, e := range effects {
		if e.Type == effect.Type && e.Target == effect.Target {
			t.effects[hostname] = append(effects[:i], effects[i+1:]...)
			return
		}
	}
}

// Get all effects for a host
func (t *EffectTracker) Get(hostname string) []ActiveEffect {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Return internal mutable state
	// Fix issue: when i added 6 ip for blackhole. in 3 hosts. But when i quit
	// it failed to clean all effects.
	original := t.effects[hostname]
	copied := make([]ActiveEffect, len(original)) // Allocate slice
	copy(copied, original)                        // Copy data into slice since copy() is just shallow copy!
	return copied
}

// GetAll, return all effects for all hosts
func (t *EffectTracker) GetAll() map[string][]ActiveEffect {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.effects
}

// Clear all effects for a host
func (t *EffectTracker) Clear(hostname string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.effects, hostname)
}

// ============================================================
// Undo Stack - session-scoped, only tracks add actions
// ============================================================

// UndoAction represents one undoable action
type UndoAction struct {
	Hostname string
	Effect   ActiveEffect // Type+Target+Value → enough to call removeSingleEffect
	Client   *ssh.Client
	BatchID  string // Group actions share same BatchID for batch undo
}

// Global undo stack
var undoStack = &UndoStack{}

// UndoStack manages undo actions with mutex for thread safety
type UndoStack struct {
	mu      sync.Mutex
	actions []UndoAction
}

// Push adds an action to the undo stack
func (s *UndoStack) Push(action UndoAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions = append(s.actions, action)
}

// Pop removes and returns the last action, returns nil if empty
func (s *UndoStack) Pop() *UndoAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.actions) == 0 {
		return nil
	}
	last := s.actions[len(s.actions)-1]
	s.actions = s.actions[:len(s.actions)-1]
	return &last
}

// Peek returns the last action without removing it, returns nil if empty
func (s *UndoStack) Peek() *UndoAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.actions) == 0 {
		return nil
	}
	return &s.actions[len(s.actions)-1]
}

// Len returns current stack size
func (s *UndoStack) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.actions)
}

// TagBatch sets BatchID on all entries from fromIndex to end
func (s *UndoStack) TagBatch(fromIndex int, batchID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := fromIndex; i < len(s.actions); i++ {
		s.actions[i].BatchID = batchID
	}
}

// PopBatch pops all actions with same BatchID as top entry.
// If top has no BatchID, pops just 1 (normal undo behavior).
func (s *UndoStack) PopBatch() []UndoAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.actions) == 0 {
		return nil
	}

	last := s.actions[len(s.actions)-1]

	// No batch ID → single undo
	if last.BatchID == "" {
		s.actions = s.actions[:len(s.actions)-1]
		return []UndoAction{last}
	}

	// Collect all actions with same BatchID
	var batch []UndoAction
	var remaining []UndoAction
	for _, a := range s.actions {
		if a.BatchID == last.BatchID {
			batch = append(batch, a)
		} else {
			remaining = append(remaining, a)
		}
	}
	s.actions = remaining
	return batch
}
