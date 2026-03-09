package authz

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/grokify/coreforge/authz"
	"github.com/grokify/coreforge/authz/simple"
	"github.com/grokify/coreforge/authz/spicedb"
	"github.com/plexusone/dashforge/ent"
	"github.com/plexusone/dashforge/ent/membership"
	"github.com/plexusone/dashforge/ent/user"
)

// Mode specifies the authorization backend.
type Mode string

const (
	// ModeSimple uses the simple role hierarchy provider.
	ModeSimple Mode = "simple"
	// ModeSpiceDB uses SpiceDB for relationship-based access control.
	ModeSpiceDB Mode = "spicedb"
)

// Role constants for DashForge.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// RoleHierarchy defines the role hierarchy for DashForge.
var RoleHierarchy = authz.RoleHierarchy{
	RoleOwner:  100,
	RoleAdmin:  80,
	RoleEditor: 60,
	RoleViewer: 40,
}

// Service provides authorization operations for DashForge.
type Service struct {
	db             *ent.Client
	mode           Mode
	simpleProvider *simple.Provider
	spiceProvider  *spicedb.Provider
	syncer         authz.RelationshipSyncer
	syncMode       authz.SyncMode
	logger         *slog.Logger
}

// ServiceOption configures the Service.
type ServiceOption func(*Service)

// WithMode sets the authorization mode.
func WithMode(mode Mode) ServiceOption {
	return func(s *Service) {
		s.mode = mode
	}
}

// WithSpiceDBProvider sets the SpiceDB provider.
func WithSpiceDBProvider(provider *spicedb.Provider) ServiceOption {
	return func(s *Service) {
		s.spiceProvider = provider
		s.syncer = spicedb.NewSyncer(provider.Client())
	}
}

// WithSyncer sets the relationship syncer.
func WithSyncer(syncer authz.RelationshipSyncer) ServiceOption {
	return func(s *Service) {
		s.syncer = syncer
	}
}

// WithSyncMode sets the synchronization mode.
func WithSyncMode(mode authz.SyncMode) ServiceOption {
	return func(s *Service) {
		s.syncMode = mode
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

// NewService creates a new authorization service.
func NewService(db *ent.Client, opts ...ServiceOption) *Service {
	s := &Service{
		db:       db,
		mode:     ModeSimple,
		syncMode: authz.SyncModeEventual,
		logger:   slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Initialize simple provider (always available for fallback)
	s.simpleProvider = simple.New(
		simple.WithRoleHierarchy(RoleHierarchy),
		simple.WithRoleGetter(s.getRoleFromDB),
		simple.WithPlatformAdminChecker(s.isPlatformAdminFromDB),
		simple.WithOwnerFullAccess(true),
		simple.WithPlatformAdminBypass(true),
	)

	return s
}

// Syncer returns the relationship syncer for identity lifecycle sync.
func (s *Service) Syncer() authz.RelationshipSyncer {
	return s.syncer
}

// Mode returns the current authorization mode.
func (s *Service) Mode() Mode {
	return s.mode
}

// provider returns the active authorization provider.
func (s *Service) provider() authz.Authorizer {
	if s.mode == ModeSpiceDB && s.spiceProvider != nil {
		return s.spiceProvider
	}
	return s.simpleProvider
}

// getRoleFromDB retrieves a user's role in an organization from the database.
func (s *Service) getRoleFromDB(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	mem, err := s.db.Membership.Query().
		Where(
			membership.UserID(userID),
			membership.OrganizationID(orgID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil // No membership = no role
		}
		return "", err
	}
	return mem.Role, nil
}

// isPlatformAdminFromDB checks if a user is a platform administrator.
func (s *Service) isPlatformAdminFromDB(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := s.db.User.Query().
		Where(user.ID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return u.IsPlatformAdmin, nil
}

// Can checks if a user can perform an action on a resource.
func (s *Service) Can(ctx context.Context, principal authz.Principal, action authz.Action, resource authz.Resource) (bool, error) {
	return s.provider().Can(ctx, principal, action, resource)
}

// CanAll checks if a user can perform all specified actions on a resource.
func (s *Service) CanAll(ctx context.Context, principal authz.Principal, actions []authz.Action, resource authz.Resource) (bool, error) {
	return s.provider().CanAll(ctx, principal, actions, resource)
}

// CanAny checks if a user can perform any of the specified actions on a resource.
func (s *Service) CanAny(ctx context.Context, principal authz.Principal, actions []authz.Action, resource authz.Resource) (bool, error) {
	return s.provider().CanAny(ctx, principal, actions, resource)
}

// Filter returns only the resources the user can access with the given action.
func (s *Service) Filter(ctx context.Context, principal authz.Principal, action authz.Action, resources []authz.Resource) ([]authz.Resource, error) {
	return s.provider().Filter(ctx, principal, action, resources)
}

// GetRole returns the user's role in an organization.
func (s *Service) GetRole(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	return s.getRoleFromDB(ctx, userID, orgID)
}

// IsMember checks if a user is a member of an organization.
func (s *Service) IsMember(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	role, err := s.getRoleFromDB(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	return role != "", nil
}

// IsPlatformAdmin checks if a user has platform-wide admin access.
func (s *Service) IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return s.isPlatformAdminFromDB(ctx, userID)
}

// CheckMinRole checks if a user has at least the specified role level.
func (s *Service) CheckMinRole(ctx context.Context, userID, orgID uuid.UUID, minRole string) (bool, error) {
	role, err := s.GetRole(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	if role == "" {
		return false, nil
	}
	return role == minRole || RoleHierarchy.CanAccess(role, minRole), nil
}

// IsOwnerOrAdmin checks if a user is owner or admin of an organization.
func (s *Service) IsOwnerOrAdmin(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	role, err := s.GetRole(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	return role == RoleOwner || role == RoleAdmin, nil
}

// Verify Service implements the required interface.
var _ authz.Authorizer = (*Service)(nil)
