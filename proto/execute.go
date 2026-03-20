package proto

import (
	"context"
	"errors"
)

// Executor sends CDP commands and receives results.
type Executor interface {
	Execute(ctx context.Context, method Method, params, result any) error
}

type executorKey struct{}

// ErrNoExecutor is returned when no Executor is found in the context.
var ErrNoExecutor = errors.New("bonk: no executor in context")

// WithExecutor returns a new context with the given Executor attached.
func WithExecutor(ctx context.Context, exec Executor) context.Context {
	return context.WithValue(ctx, executorKey{}, exec)
}

// ExecutorFromContext returns the Executor attached to the context.
func ExecutorFromContext(ctx context.Context) (Executor, error) {
	exec, ok := ctx.Value(executorKey{}).(Executor)
	if !ok || exec == nil {
		return nil, ErrNoExecutor
	}
	return exec, nil
}

// Execute sends a CDP command using the Executor from the context.
func Execute(ctx context.Context, method Method, params, result any) error {
	exec, err := ExecutorFromContext(ctx)
	if err != nil {
		return err
	}
	return exec.Execute(ctx, method, params, result)
}
