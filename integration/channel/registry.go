package channel

import (
	"sort"
	"sync"
)

var (
	channelsMu sync.RWMutex
	channels   = make(map[string]Channel)
)

// Register adds a channel to the registry.
// It should be called from init() in channel packages.
// If a channel with the same type already exists, it is replaced.
func Register(c Channel) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	channels[c.Type()] = c
}

// Get retrieves a channel by type.
// Returns the channel and true if found, nil and false otherwise.
func Get(channelType string) (Channel, bool) {
	channelsMu.RLock()
	defer channelsMu.RUnlock()
	c, ok := channels[channelType]
	return c, ok
}

// MustGet retrieves a channel by type or panics if not found.
func MustGet(channelType string) Channel {
	c, ok := Get(channelType)
	if !ok {
		panic("channel: unknown channel type " + channelType)
	}
	return c
}

// Available returns all registered channel types in sorted order.
func Available() []string {
	channelsMu.RLock()
	defer channelsMu.RUnlock()

	types := make([]string, 0, len(channels))
	for t := range channels {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// Registered returns a copy of all registered channels.
func Registered() map[string]Channel {
	channelsMu.RLock()
	defer channelsMu.RUnlock()

	result := make(map[string]Channel, len(channels))
	for t, c := range channels {
		result[t] = c
	}
	return result
}

// Unregister removes a channel from the registry.
// This is primarily useful for testing.
func Unregister(channelType string) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	delete(channels, channelType)
}

// Reset clears all registered channels.
// This is primarily useful for testing.
func Reset() {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	channels = make(map[string]Channel)
}

// ChannelInfo provides metadata about a registered channel.
type ChannelInfo struct {
	Type         string       `json:"type"`
	Name         string       `json:"name"`
	Capabilities Capabilities `json:"capabilities"`
}

// ListChannels returns information about all registered channels.
func ListChannels() []ChannelInfo {
	channelsMu.RLock()
	defer channelsMu.RUnlock()

	infos := make([]ChannelInfo, 0, len(channels))
	for _, c := range channels {
		infos = append(infos, ChannelInfo{
			Type:         c.Type(),
			Name:         c.Name(),
			Capabilities: c.Capabilities(),
		})
	}

	// Sort by type for consistent ordering
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Type < infos[j].Type
	})

	return infos
}
