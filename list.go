package main

import (
	"fmt"
)

// List interface.
type List interface {
	Len() int
	Front() *listItem
	Back() *listItem
	PushFront(v interface{}) *listItem
	PushBack(v interface{}) *listItem
	Remove(i *listItem)
	MoveToFront(i *listItem)

	display()
}

type listItem struct {
	Value interface{}
	Prev  *listItem
	Next  *listItem
}

type list struct {
	len   int
	front *listItem
	back  *listItem
}

// NewList - create new List.
func NewList() List {
	return &list{}
}

// Len - return length of list.
func (l *list) Len() int {
	return l.len
}

// Just for debugging purpose.
func (l *list) display() {
	fmt.Print("Displaying the list:", l)
	fmt.Print(" Items:")
	for i := l.Front(); i != nil; i = i.Next {
		fmt.Print(i)
	}
	fmt.Println("")
}

// Front - return first element of list.
func (l *list) Front() *listItem {
	return l.front
}

// Back - return last element of list.
func (l *list) Back() *listItem {
	return l.back
}

// PushFront - add value to the beginning of the list.
func (l *list) PushFront(v interface{}) *listItem {
	item := &listItem{Value: v}
	// Correct linking for previous front item
	item.Next = l.front
	if l.front != nil {
		l.front.Prev = item
	}
	l.front = item
	// Need to set opposite item to current one, when first element was added
	if l.Len() == 0 {
		l.back = item
	}
	l.len++
	return item
}

// PushBack - add value to the end of the list.
func (l *list) PushBack(v interface{}) *listItem {
	item := &listItem{Value: v}
	// Correct linking for previous back item
	item.Prev = l.back
	if l.back != nil {
		l.back.Next = item
	}
	l.back = item
	// Need to set opposite item to current one, when first element was added
	if l.Len() == 0 {
		l.front = item
	}
	l.len++
	return item
}

// Remove - remove item from the list.
func (l *list) Remove(item *listItem) {
	// Remove first item
	if item.Prev == nil {
		l.front = item.Next
	} else {
		item.Prev.Next = item.Next
	}
	// Remove last item
	if item.Next == nil {
		l.back = item.Prev
	} else {
		item.Next.Prev = item.Prev
	}
	l.len--
}

// MoveToFront - move item at the beginning of the list.
func (l *list) MoveToFront(item *listItem) {
	// Skip doing anything, if we try to move the first item
	if item.Prev == nil {
		return
	}
	if item.Next == nil {
		// Update link for last element
		l.back = l.back.Prev
		l.back.Next = nil
	} else {
		// Update link of next element
		item.Next.Prev = item.Prev
	}
	item.Prev = nil
	item.Next = l.front
	// Update links for front element
	l.front.Prev = item
	l.front = item
}
