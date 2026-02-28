# Linear API Documentation Snapshots

This directory contains markdown snapshots of Linear's API documentation, organized by topic for easy reference.

## Scope and Intent

- These files document **Linear's API capabilities**, not the exact current feature surface of `linctl`.
- Some topics here (for example, webhooks or OAuth app flows) may be useful background for future CLI work but are not exposed as top-level commands today.
- Treat these as offline API reference material; use `linctl --help` and `README.md` for current CLI behavior.

## Current `linctl` Coverage (This Branch)

- `issue`: list/search/get/assign/create/update
  - Includes project linking (`--project`), project milestone linking (`--project-milestone`), and parent-child linking (`--parent`) on update flows.
- `project`: list/get/create/update/delete (archive by default, permanent with `--permanent`)
- `comment`: list/create
- `team`: list/get/members
- `user`: list/get/me
- `auth`: login/status/logout

## Core API Documentation

### [GraphQL API](./graphql-api.md)
- Main API endpoint and basics
- Authentication methods (OAuth2 and Personal API Keys)
- Basic query and mutation examples
- Best practices and support information

### [Authentication](./oauth-authentication.md)
- OAuth 2.0 flow setup and configuration
- Authorization request parameters and scopes
- Token exchange process
- Making authenticated API requests

### [Rate Limiting](./rate-limiting.md)
- Request limits for different authentication types
- Complexity limits and headers
- Best practices for avoiding rate limits
- Error handling for rate limit exceeded

## Data Management

### [Pagination](./pagination.md)
- Cursor-based pagination model
- Pagination arguments and best practices
- Alternative syntax options
- Ordering results by creation or update time

### [Filtering](./filtering.md)
- Available comparators for different data types
- Logical operators (AND/OR)
- Query examples for complex filtering
- Relative time and null value filtering

### [Attachments](./attachments.md)
- Creating and updating attachments
- Metadata support and rich features
- API examples and use cases
- Querying attachments by URL

## File Operations

### [File Upload](./file-upload.md)
- Multiple methods for including files
- Manual upload process with pre-signed URLs
- Server-side upload examples
- Common errors and supported file types

### [File Storage Authentication](./file-storage-authentication.md)
- Accessing stored files with authentication
- Signed URLs for temporary access
- Security considerations

## Integration Features

### [Webhooks](./webhooks.md)
- Webhook setup and configuration
- Supported models and event types
- Payload structure and security
- Example webhook consumer implementation

### [TypeScript SDK](./typescript-sdk.md)
- Installation and setup
- Authentication methods
- Basic usage examples with async/await and promises
- Error handling and key features

## API Lifecycle

### [Deprecations](./deprecations.md)
- Linear's approach to API versioning
- Deprecation mechanism using GraphQL directives
- Migration strategies and best practices
- How to stay updated on changes

## Quick Reference

### Authentication Headers
```bash
# Personal API Key
Authorization: <API_KEY>

# OAuth Token
Authorization: Bearer <ACCESS_TOKEN>
```

### API Endpoints
- **GraphQL API**: `https://api.linear.app/graphql`
- **OAuth Token**: `https://api.linear.app/oauth/token`
- **File Storage**: `https://uploads.linear.app`

### Key Resources
- [Apollo Studio Schema Explorer](https://studio.apollographql.com/public/Linear-API/schema/reference?variant=current)
- [Linear Changelog](https://linear.app/changelog) (look for `[API]` entries)
- [Linear Developer Portal](https://linear.app/developers)

---

**Last Updated**: July 13, 2025  
**Source**: https://linear.app/developers

These snapshots provide offline reference material for Linear's API documentation. For the most current information, always refer to the official Linear developer documentation.
