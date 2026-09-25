// Package operation provides implementation for handling operations in Yandex Cloud Go SDK.
//
// An operation represents a long-running task in Yandex Cloud that can be monitored
// and managed through this package. Operations are typically created by service methods
// and can be polled for completion status.
//
// Usage Example:
//
//	op, err := service.CreateResource(ctx, request)
//	if err != nil {
//	    // handle error
//	}
//
//	// Wait for operation completion
//	result, err := op.Wait(ctx)
//	if err != nil {
//	    // handle error
//	}
package operation
