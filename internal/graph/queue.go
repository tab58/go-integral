package graph

type ListItem[T any] struct {
	Value T
	Next  *ListItem[T]
}

type Queue[T any] struct {
	Head *ListItem[T]
	Tail *ListItem[T]
}

func (q *Queue[T]) Enqueue(item T) {
	if q.Head == nil {
		q.Head = &ListItem[T]{
			Value: item,
		}
		q.Tail = q.Head
	} else {
		q.Tail.Next = &ListItem[T]{
			Value: item,
		}
		q.Tail = q.Tail.Next
	}
}

func (q *Queue[T]) Dequeue() *T {
	if q.Head == nil {
		return nil
	}
	item := q.Head.Value

	q.Head = q.Head.Next
	if q.Head == nil {
		q.Tail = nil
	}

	return &item
}

func (q *Queue[T]) IsEmpty() bool {
	return q.Head == nil
}
