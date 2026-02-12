package main

import "sync"

// Global tracker instance
var effectTracker = NewEffectTracker() // Global variable, accessible from anywhere

// Effect types
const (
	EffectBlackHole  = "blackhole"
	EffectLatency    = "latency"
	EffectPacketLoss = "packetloss"
	EffectPartition  = "partition"
	EffectPortBlock  = "portblock"
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
