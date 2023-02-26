package event

import (
	"reflect"
	"sort"
)

// Listener interface
type Listener interface {
	Handle(e Event) error
}

// ListenerFunc func definition.
type ListenerFunc func(e Event) error

// Handle event. implements the Listener interface
func (fn ListenerFunc) Handle(e Event) error {
	return fn(e)
}

// Subscriber event subscriber interface.
// you can register multi event listeners in a struct func.
type Subscriber interface {
	// SubscribedEvents register event listeners
	// key: is event name
	// value: can be Listener or ListenerItem interface
	SubscribedEvents() map[string]any
}

// ListenerItem storage a event listener and it's priority value.
type ListenerItem struct {
	Priority int
	Listener Listener
}

/*************************************************************
 * Listener Queue
 *************************************************************/

// ListenerQueue storage sorted Listener instance.
type ListenerQueue struct {
	items []*ListenerItem
}

// Len get items length
func (lq *ListenerQueue) Len() int {
	return len(lq.items)
}

// IsEmpty get items length == 0
func (lq *ListenerQueue) IsEmpty() bool {
	return len(lq.items) == 0
}

// Push get items length
func (lq *ListenerQueue) Push(li *ListenerItem) *ListenerQueue {
	lq.items = append(lq.items, li)
	return lq
}

// Sort the queue items by ListenerItem's priority.
// Priority:
//
//	High > Low
func (lq *ListenerQueue) Sort() *ListenerQueue {
	// if lq.IsEmpty() {
	// 	return lq
	// }
	ls := ByPriorityItems(lq.items)

	// check items is sorted
	if !sort.IsSorted(ls) {
		sort.Sort(ls)
	}

	return lq
}

// Items get all ListenerItem
func (lq *ListenerQueue) Items() []*ListenerItem {
	return lq.items
}

// Remove a listener from the queue
func (lq *ListenerQueue) Remove(listener Listener) {
	if listener == nil {
		return
	}

	// unsafe.Pointer(listener)
	ptrVal := getListenCompareKey(listener)

	var newItems []*ListenerItem
	for _, li := range lq.items {
		liPtrVal := getListenCompareKey(li.Listener)
		if liPtrVal == ptrVal {
			continue
		}

		newItems = append(newItems, li)
	}

	lq.items = newItems
}

// Clear all listeners
func (lq *ListenerQueue) Clear() {
	lq.items = lq.items[:0]
}

/////
func getListenCompareKey(src Listener) reflect.Value {
	return reflect.ValueOf(src)
}

/*************************************************************
 * Sorted PriorityItems
 *************************************************************/

// ByPriorityItems type. implements the sort.Interface
type ByPriorityItems []*ListenerItem

// Len get items length
func (ls ByPriorityItems) Len() int {
	return len(ls)
}

// Less implements the sort.Interface.Less.
func (ls ByPriorityItems) Less(i, j int) bool {
	return ls[i].Priority > ls[j].Priority
}

// Swap implements the sort.Interface.Swap.
func (ls ByPriorityItems) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}
