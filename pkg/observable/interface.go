package observable

// TODO(@bryanchriswhite): Add a README for the observable.

import "context"

// NOTE: We explicitly decided to write a small and custom notifications package
// to keep logic simple and minimal. If the needs & requirements of this library ever
// grow, other packages (e.g. https://github.com/ReactiveX/RxGo) can be considered.
// (see: https://github.com/ReactiveX/RxGo/pull/377)

// Observable is a generic interface that allows multiple subscribers to be
// notified of new values asynchronously.
// It is analogous to a publisher in a "Fan-Out" system design.
type Observable[V any] interface {
	Subscribe(context.Context) Observer[V]
	Close()
}

// Observer is a generic interface that provides access to the notified
// channel and allows unsubscribing from an Observable.
// It is analogous to a subscriber in a "Fan-Out" system design.
type Observer[V any] interface {
	Unsubscribe()
	Ch() <-chan V
}
