package server

import "context"

type Observer[T any] struct {
	ctx    context.Context
	values chan T
	cancel context.CancelFunc
}

func NewObserver[T any](ctx context.Context) *Observer[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &Observer[T]{
		ctx:    ctx,
		values: make(chan T),
		cancel: cancel,
	}
}

func (o *Observer[T]) Send(value T) {
	select {
	case <-o.ctx.Done():
		return
	case o.values <- value:
	}
}

func (o *Observer[T]) Range() <-chan T {
	results := make(chan T)

	go func() {
		defer close(results)
		for {
			select {
			case <-o.ctx.Done():
				return
			case value, ok := <-o.values:
				if !ok {
					return
				}
				results <- value
			}
		}
	}()

	return results
}

func (o *Observer[T]) Values() <-chan T {
	return o.values
}

func (o *Observer[T]) Close() {
	o.cancel()
}

func (o *Observer[T]) IsCanceled() bool {
	return o.ctx.Err() != nil
}
