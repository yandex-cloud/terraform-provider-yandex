// Package transport provides functionality for managing gRPC connections and communication
// in the Yandex Cloud Go SDK. It includes interfaces and implementations for connection
// management, endpoint resolution, and connection pooling.
//
// The package contains the following subpackages:
//   - grpc: Implements gRPC-specific functionality including connection pooling and client management
//   - middleware: Provides interceptors and middleware components for extending gRPC functionality
//
// Key components in the main package:
//   - Connector: Interface for retrieving gRPC client connections
//   - ConnectorImpl: Implementation of the Connector interface with connection pooling
//   - SingleConnector: Simplified connector for single endpoint scenarios
package transport
