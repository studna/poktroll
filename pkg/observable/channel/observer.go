package channel

import (
	"context"
	"sync"
	"time"

	"pocket/pkg/observable"
)

const (
	// TODO_DISCUSS: what should this be? should it be configurable? It seems to be most
	// relevant in the context of the behavior of the observable when it has multiple
	// observers which consume at different rates.
	// observerBufferSize is the buffer size of a channelObserver's channel.
	observerBufferSize = 1
	// sendRetryInterval is the duration between attempts to send on the observer's
	// channel in the event that it's full. It facilitates a branch in a for loop
	// which unlocks the observer's mutex and tries again.
	// NOTE: setting this too low can cause the send retry loop to "slip", giving
	// up on a send attempt before the channel is ready to receive for multiple
	// iterations of the loop.
	sendRetryInterval = 100 * time.Millisecond
)

var _ observable.Observer[any] = &channelObserver[any]{}

// channelObserver implements the observable.Observer interface.
type channelObserver[V any] struct {
	ctx context.Context
	// onUnsubscribe is called in Observer#Unsubscribe, removing the respective
	// observer from observers in a concurrency-safe manner.
	onUnsubscribe func(toRemove *channelObserver[V])
	// observerMu protects the observerCh and isClosed fields.
	observerMu *sync.RWMutex
	// observerCh is the channel that is used to emit values to the observer.
	// I.e. on the "N" side of the 1:N relationship between observable and
	// observer.
	observerCh chan V
	// isClosed indicates whether the observer has been isClosed. It's set in
	// unsubscribe; isClosed observers can't be reused.
	isClosed bool
}

type UnsubscribeFunc[V any] func(toRemove *channelObserver[V])

func NewObserver[V any](
	ctx context.Context,
	onUnsubscribe UnsubscribeFunc[V],
) *channelObserver[V] {
	// Create a channel for the observer and append it to the observers list
	return &channelObserver[V]{
		ctx:           ctx,
		observerMu:    new(sync.RWMutex),
		observerCh:    make(chan V, observerBufferSize),
		onUnsubscribe: onUnsubscribe,
	}
}

// Unsubscribe closes the subscription channel and removes the subscription from
// the observable.
func (obsvr *channelObserver[V]) Unsubscribe() {
	obsvr.unsubscribe()
}

// Ch returns a receive-only subscription channel.
func (obsvr *channelObserver[V]) Ch() <-chan V {
	return obsvr.observerCh
}

// unsubscribe closes the subscription channel, marks the observer as isClosed, and
// removes the subscription from its observable's observers list via onUnsubscribe.
func (obsvr *channelObserver[V]) unsubscribe() {
	obsvr.observerMu.Lock()
	defer obsvr.observerMu.Unlock()

	if obsvr.isClosed {
		return
	}

	close(obsvr.observerCh)
	obsvr.isClosed = true
	obsvr.onUnsubscribe(obsvr)
}

// notify is used called by observable to send on the observer channel. Can't
// use channelObserver#Ch because it's receive-only. It will block if the channel
// is full but will release the read-lock for half of the sendRetryInterval. The
// other half holds is spent holding the read-lock and waiting for the (full)
// channel to be ready to receive.
func (obsvr *channelObserver[V]) notify(value V) {
	defer obsvr.observerMu.RUnlock()

	sendRetryTicker := time.NewTicker(sendRetryInterval / 2)
	for {
		// observerMu must remain read-locked until the value is sent on observerCh
		// in the event that it would be isClosed concurrently (i.e. this observer
		// unsubscribes), which could cause a "send on isClosed channel" error.
		if !obsvr.observerMu.TryRLock() {
			time.Sleep(sendRetryInterval / 2)
			continue
		}
		if obsvr.isClosed {
			return
		}

		select {
		case <-obsvr.ctx.Done():
			// if the context is done just release the read-lock (deferred)
			return
		case obsvr.observerCh <- value:
			return
		// if the context isn't done and channel is full (i.e. blocking),
		// release the read-lock to give write-lockers a turn. This case
		// continues the loop, re-read-locking and trying again.
		case <-sendRetryTicker.C:
			// CONSIDERATION: repurpose this retry loop into a default path which
			// buffers values so that the publishCh doesn't block and other observers
			// can still be notified.
			// TECHDEBT: add some logic to drain the buffer at some appropriate time

			// this case implies that the (read) lock was acquired, so it must
			// be unlocked before continuing the send retry loop.
			obsvr.observerMu.RUnlock()
		}
		time.Sleep(sendRetryInterval / 2)
	}
}
