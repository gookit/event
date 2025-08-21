package event

import (
	"context"
	"fmt"
	"strings"
)

/*************************************************************
 * region Trigger event
 *************************************************************/

// MustTrigger alias of the method MustFire()
func (em *Manager) MustTrigger(name string, params M) Event { return em.MustFire(name, params) }

// MustFire fire event by name. will panic on error
func (em *Manager) MustFire(name string, params M) Event {
	err, e := em.Fire(name, params)
	if err != nil {
		panic(err)
	}
	return e
}

// Trigger alias of the method Fire()
func (em *Manager) Trigger(name string, params M) (error, Event) { return em.Fire(name, params) }

// Fire trigger event by name. if not found listener, will return (nil, nil)
func (em *Manager) Fire(name string, params M) (err error, e Event) {
	// call listeners handle event
	e, err = em.fireByNameCtx(em.ctx, name, params, false)
	return
}

// FireCtx fire event by name with context
func (em *Manager) FireCtx(ctx context.Context, name string, params M) (err error, e Event) {
	// call listeners handle event
	e, err = em.fireByNameCtx(ctx, name, params, false)
	return
}

// Async fire event by go channel.
//
// Note: if you want to use this method, you should
// call the method Close() after all events are fired.
func (em *Manager) Async(name string, params M) {
	_, _ = em.fireByName(name, params, true)
}

// FireC async fire event by go channel. alias of the method Async()
//
// Note: if you want to use this method, you should
// call the method Close() after all events are fired.
func (em *Manager) FireC(name string, params M) {
	_, _ = em.fireByName(name, params, true)
}

// fire event by name.
//
// if useCh is true, will async fire by channel. always return (nil, nil)
//
// On useCh=false:
//   - will call listeners handle event.
//   - if not found listener, will return (nil, nil)
func (em *Manager) fireByName(name string, params M, useCh bool) (Event, error) {
	return em.fireByNameCtx(em.ctx, name, params, useCh)
}

// fireByNameCtx fire event by name with context
func (em *Manager) fireByNameCtx(ctx context.Context, name string, params M, useCh bool) (e Event, err error) {
	name, err = goodNameOrErr(name, false)
	if err != nil {
		return nil, err
	}

	// use pre-defined Event
	if fc, ok := em.eventFc[name]; ok {
		e = fc() // make new instance
		if params != nil {
			e.SetData(params)
		}
	} else {
		// create new basic event instance
		e = em.newBasicEvent(name, params)
	}

	// fire by channel
	if useCh {
		em.FireAsync(e)
		return nil, nil
	}

	// call listeners handle event
	err = em.FireEventCtx(ctx, e)
	return
}

// FireEvent fire event by given Event instance
func (em *Manager) FireEvent(e Event) error {
	if ec, ok := e.(ContextAble); ok {
		return em.FireEventCtx(ec.Context(), e)
	}
	return em.FireEventCtx(em.ctx, e)
}

// FireEventCtx fire event by given Event instance with context
func (em *Manager) FireEventCtx(ctx context.Context, e Event) (err error) {
	if em.EnableLock {
		em.Lock()
		defer em.Unlock()
	}

	// ensure aborted is false.
	e.Abort(false)
	name := e.Name()
	// set context
	if ec, ok := e.(ContextAble); ok {
		ec.WithContext(ctx)
	}

	// fire group listeners by wildcard. eg "db.user.*"
	if em.MatchMode == ModePath {
		err = em.firePathMode(ctx, name, e)
		return
	}

	// handle mode: ModeSimple
	err = em.fireSimpleMode(ctx, name, e)
	if err != nil || e.IsAborted() {
		return
	}

	// fire wildcard event listeners
	if lq, ok := em.listeners[Wildcard]; ok {
		for _, li := range lq.Sort().Items() {
			// Check context cancellation
			if ctx != nil {
				select {
				case <-ctx.Done():
					err = ctx.Err()
					return
				default:
				}
			}

			err = li.Listener.Handle(e)
			if err != nil || e.IsAborted() {
				break
			}
		}
	}
	return
}

// ModeSimple has group listeners by wildcard. eg "db.user.*"
//
// Example:
//   - event "db.user.add" will trigger listeners on the "db.user.*"
func (em *Manager) fireSimpleMode(ctx context.Context, name string, e Event) (err error) {
	// fire direct matched listeners. eg: db.user.add
	if lq, ok := em.listeners[name]; ok {
		// sort by priority before call.
		for _, li := range lq.Sort().Items() {
			// Check context cancellation
			if ctx != nil {
				select {
				case <-ctx.Done():
					err = ctx.Err()
					return
				default:
				}
			}

			err = li.Listener.Handle(e)
			if err != nil || e.IsAborted() {
				return
			}
		}
	}

	pos := strings.LastIndexByte(name, '.')

	// exists group
	if pos > 0 && pos < len(name) {
		groupName := name[:pos+1] + Wildcard // "app.*"

		if lq, ok := em.listeners[groupName]; ok {
			for _, li := range lq.Sort().Items() {
				// Check context cancellation
				if ctx != nil {
					select {
					case <-ctx.Done():
						err = ctx.Err()
						return
					default:
					}
				}

				err = li.Listener.Handle(e)
				if err != nil || e.IsAborted() {
					return
				}
			}
		}
	}

	return nil
}

// firePathMode fire group listeners by ModePath.
//
// Example:
//   - event "db.user.add" will trigger listeners on the "db.**"
//   - event "db.user.add" will trigger listeners on the "db.user.*"
func (em *Manager) firePathMode(ctx context.Context, name string, e Event) (err error) {
	for pattern, lq := range em.listeners {
		if pattern == name || matchNodePath(pattern, name, ".") {
			for _, li := range lq.Sort().Items() {
				// Check context cancellation
				if ctx != nil {
					select {
					case <-ctx.Done():
						err = ctx.Err()
						return
					default:
					}
				}

				err = li.Listener.Handle(e)
				if err != nil || e.IsAborted() {
					return
				}
			}
		}
	}

	return nil
}

/*************************************************************
 * region Fire by channel
 *************************************************************/

// FireAsync async fire event by go channel.
//
// Note: if you want to use this method, you should
// call the method Close() after all events are fired.
//
// Example:
//
//	em := NewManager("test")
//	em.FireAsync("db.user.add", M{"id": 1001})
func (em *Manager) FireAsync(e Event) {
	// once make consumers
	em.oc.Do(func() {
		em.makeConsumers()
	})

	// dispatch event
	em.ch <- e
}

// async fire event by 'go' keywords
func (em *Manager) makeConsumers() {
	if em.ConsumerNum <= 0 {
		em.ConsumerNum = defaultConsumerNum
	}
	if em.ChannelSize <= 0 {
		em.ChannelSize = defaultChannelSize
	}

	em.ch = make(chan Event, em.ChannelSize)

	// make event consumers
	for i := 0; i < em.ConsumerNum; i++ {
		em.wg.Add(1)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					em.err = fmt.Errorf("async consum event error: %v", err)
				}
				em.wg.Done()
			}()

			// keep running until channel closed
			for e := range em.ch {
				_ = em.FireEvent(e) // ignore async fire error
			}
		}()
	}
}

// FireBatch fire multi event at once.
//
// Usage:
//
//	FireBatch("name1", "name2", &MyEvent{})
func (em *Manager) FireBatch(es ...any) (ers []error) {
	var err error
	for _, e := range es {
		if name, ok := e.(string); ok {
			err, _ = em.Fire(name, nil)
		} else if evt, ok := e.(Event); ok {
			err = em.FireEvent(evt)
		} // ignore invalid param.

		if err != nil {
			ers = append(ers, err)
		}
	}
	return
}

// AsyncFire simple async fire event by 'go' keywords
func (em *Manager) AsyncFire(e Event) {
	go func(e Event) {
		_ = em.FireEvent(e)
	}(e)
}

// AwaitFire async fire event by 'go' keywords, but will wait return result
func (em *Manager) AwaitFire(e Event) (err error) {
	ch := make(chan error)

	go func(e Event) {
		err := em.FireEvent(e)
		ch <- err
	}(e)

	err = <-ch
	close(ch)
	return
}

/*************************************************************
 * region Helper methods
 *************************************************************/

// CloseWait close channel and wait all async event done.
func (em *Manager) CloseWait() error {
	if err := em.Close(); err != nil {
		return err
	}
	return em.Wait()
}

// Wait wait all async event done.
func (em *Manager) Wait() error {
	em.wg.Wait()
	return em.err
}
