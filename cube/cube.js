// Cube.js configuration for Dashforge
// See: https://cube.dev/docs/config

module.exports = {
  // Database configuration (uses environment variables)
  // CUBEJS_DB_TYPE, CUBEJS_DB_HOST, CUBEJS_DB_NAME, etc.

  // Enable scheduled refresh for pre-aggregations
  scheduledRefreshTimer: 60,

  // API settings
  contextToAppId: ({ securityContext }) => {
    // Support multi-tenant isolation
    return securityContext?.tenant_id || 'default';
  },

  // Enable GraphQL (optional)
  // graphQL: {
  //   enabled: true,
  // },

  // Pre-aggregations configuration
  preAggregationsSchema: ({ securityContext }) => {
    return 'pre_aggregations';
  },

  // Security context from JWT
  checkAuth: (req, auth) => {
    // Example: Extract tenant from JWT token
    // const decoded = jwt.verify(auth, process.env.JWT_SECRET);
    // return { tenant_id: decoded.tenant_id };
    return {};
  },

  // Query rewrite (for RLS)
  queryRewrite: (query, { securityContext }) => {
    // Example: Add tenant filter
    // if (securityContext?.tenant_id) {
    //   query.filters = query.filters || [];
    //   query.filters.push({
    //     member: 'TenantBase.tenantId',
    //     operator: 'equals',
    //     values: [securityContext.tenant_id]
    //   });
    // }
    return query;
  },
};
