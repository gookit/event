package event

import "sort"

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

// ListenerQueue storage sorted Listener instance.
type ListenerQueue struct {
	items []*ListenerItem
}

// Len get items length
func (lq *ListenerQueue) Len() int {
	return len(lq.items)
}

// Len get items length
func (lq *ListenerQueue) Push(li *ListenerItem) *ListenerQueue {
	lq.items = append(lq.items, li)
	return lq
}

// Sort the queue items by ListenerItem's priority
func (lq *ListenerQueue) Sort() *ListenerQueue {
	ls := ListenerItems(lq.items)

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

// Clear clear all listeners
func (lq *ListenerQueue) Clear() {
	lq.items = lq.items[0:0]
}

// ListenerItem storage a event listener and it's priority value.
type ListenerItem struct {
	priority int
	listener Listener
}

// ListenerItems type. implements the sort.Interface
type ListenerItems []*ListenerItem

// Len get items length
func (ls ListenerItems) Len() int {
	return len(ls)
}

// Less implements the sort.Interface.Less.
func (ls ListenerItems) Less(i, j int) bool {
	return ls[i].priority < ls[j].priority
}

// Swap implements the sort.Interface.Swap.
func (ls ListenerItems) Swap(i, j int) {
	ls[i].priority, ls[j].priority = ls[j].priority, ls[i].priority
}
