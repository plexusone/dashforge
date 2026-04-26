package datasource

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Manager manages connections to multiple data sources.
// It handles connection pooling, caching, and lifecycle management.
type Manager struct {
	mu          sync.RWMutex
	connections map[int]*managedConnection
	logger      *slog.Logger

	// customProviders allows injection of custom providers (omniX pattern).
	customProviders map[string]Provider
}

// managedConnection wraps a connection with metadata.
type managedConnection struct {
	conn      Connection
	config    ConnectionConfig
	provider  Provider
	createdAt time.Time
	lastUsed  time.Time
}

// ManagerConfig configures the Manager.
type ManagerConfig struct {
	// Logger for logging connection events.
	Logger *slog.Logger
}

// NewManager creates a new connection Manager.
func NewManager(cfg ManagerConfig) *Manager {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return &Manager{
		connections:     make(map[int]*managedConnection),
		customProviders: make(map[string]Provider),
		logger:          logger,
	}
}

// RegisterCustomProvider registers a custom provider (omniX pattern).
// Custom providers take precedence over built-in registered providers.
func (m *Manager) RegisterCustomProvider(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customProviders[p.Name()] = p
}

// DataSourceConfig represents a data source configuration that can be used
// to establish a connection. This is typically loaded from the database (ent.DataSource).
type DataSourceConfig struct {
	ID                  int
	Name                string
	Slug                string
	Type                string
	ConnectionURL       string
	ConnectionURLEnv    string
	MaxConnections      int
	QueryTimeoutSeconds int
	ReadOnly            bool
	SSLEnabled          bool
}

// GetConnection returns a connection for a data source, creating if needed.
func (m *Manager) GetConnection(ctx context.Context, ds DataSourceConfig) (Connection, error) {
	m.mu.RLock()
	if mc, ok := m.connections[ds.ID]; ok {
		mc.lastUsed = time.Now()
		m.mu.RUnlock()
		return mc.conn, nil
	}
	m.mu.RUnlock()

	// Create new connection
	return m.createConnection(ctx, ds)
}

// createConnection creates a new connection for a data source.
func (m *Manager) createConnection(ctx context.Context, ds DataSourceConfig) (Connection, error) {
	// Get provider (custom first, then registered)
	provider := m.getProvider(ds.Type)
	if provider == nil {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, ds.Type)
	}

	// Build connection config
	cfg := m.buildConfig(ds)

	// Validate config
	if err := provider.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	// Connect
	conn, err := provider.Connect(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Cache the connection
	m.mu.Lock()
	m.connections[ds.ID] = &managedConnection{
		conn:      conn,
		config:    cfg,
		provider:  provider,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
	}
	m.mu.Unlock()

	m.logger.Info("connection established",
		"datasource_id", ds.ID,
		"datasource_name", ds.Name,
		"provider", ds.Type,
	)

	return conn, nil
}

// getProvider returns the provider for a given type.
func (m *Manager) getProvider(typeName string) Provider {
	// Check custom providers first (omniX pattern)
	if p, ok := m.customProviders[typeName]; ok {
		return p
	}

	// Fall back to registered providers
	p, _ := Get(typeName)
	return p
}

// buildConfig builds a ConnectionConfig from a DataSourceConfig.
func (m *Manager) buildConfig(ds DataSourceConfig) ConnectionConfig {
	sslMode := ""
	if ds.SSLEnabled {
		sslMode = "require"
	} else {
		sslMode = "disable"
	}

	return ConnectionConfig{
		ConnectionURL:    ds.ConnectionURL,
		ConnectionURLEnv: ds.ConnectionURLEnv,
		MaxConnections:   ds.MaxConnections,
		QueryTimeout:     time.Duration(ds.QueryTimeoutSeconds) * time.Second,
		ReadOnly:         ds.ReadOnly,
		SSLMode:          sslMode,
	}
}

// TestConnection tests a data source connection without caching.
func (m *Manager) TestConnection(ctx context.Context, ds DataSourceConfig) error {
	// Get provider
	provider := m.getProvider(ds.Type)
	if provider == nil {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, ds.Type)
	}

	// Build config
	cfg := m.buildConfig(ds)

	// Validate
	if err := provider.ValidateConfig(cfg); err != nil {
		return err
	}

	// Connect and ping
	conn, err := provider.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	return conn.Ping(ctx)
}

// CloseConnection closes a specific data source connection.
func (m *Manager) CloseConnection(dsID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mc, ok := m.connections[dsID]
	if !ok {
		return nil // Not an error if not found
	}

	err := mc.conn.Close()
	delete(m.connections, dsID)

	m.logger.Info("connection closed", "datasource_id", dsID)

	return err
}

// CloseAll closes all managed connections.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for id, mc := range m.connections {
		if err := mc.conn.Close(); err != nil {
			lastErr = err
			m.logger.Error("error closing connection",
				"datasource_id", id,
				"error", err,
			)
		}
	}

	m.connections = make(map[int]*managedConnection)
	m.logger.Info("all connections closed")

	return lastErr
}

// Stats returns statistics for all connections.
func (m *Manager) Stats() map[int]ConnectionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[int]ConnectionStats, len(m.connections))
	for id, mc := range m.connections {
		result[id] = mc.conn.Stats()
	}
	return result
}

// ConnectionInfo returns info about a managed connection.
type ConnectionInfo struct {
	ID        int             `json:"id"`
	Provider  string          `json:"provider"`
	CreatedAt time.Time       `json:"createdAt"`
	LastUsed  time.Time       `json:"lastUsed"`
	Stats     ConnectionStats `json:"stats"`
}

// Info returns information about all managed connections.
func (m *Manager) Info() []ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ConnectionInfo, 0, len(m.connections))
	for id, mc := range m.connections {
		result = append(result, ConnectionInfo{
			ID:        id,
			Provider:  mc.provider.Name(),
			CreatedAt: mc.createdAt,
			LastUsed:  mc.lastUsed,
			Stats:     mc.conn.Stats(),
		})
	}
	return result
}

// RefreshConnection closes and reopens a connection.
func (m *Manager) RefreshConnection(ctx context.Context, ds DataSourceConfig) (Connection, error) {
	// Close existing if any
	_ = m.CloseConnection(ds.ID)

	// Create new
	return m.createConnection(ctx, ds)
}
