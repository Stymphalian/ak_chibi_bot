package misc

type RollingArray[T any] struct {
	Values []T
	start  int
	end    int
	size   int
	len    int
}

func NewRollingArray[T any](size int) *RollingArray[T] {
	return &RollingArray[T]{
		Values: make([]T, size),
		start:  0,
		end:    0,
		size:   size,
		len:    0,
	}
}

func (r *RollingArray[T]) Add(value T) {
	r.Values[r.end] = value
	r.end = (r.end + 1) % r.size
	if r.len == r.size {
		r.start = (r.start + 1) % r.size
	} else {
		r.len += 1
	}
}

func (r *RollingArray[T]) Get(index int) T {
	return r.Values[(r.start+index)%r.size]
}

func (r *RollingArray[T]) RemoveFirst() {
	if r.len == 0 {
		return
	}
	r.start = (r.start + 1) % r.size
	r.len -= 1
}

func (r *RollingArray[T]) RemoveLast() {
	if r.len == 0 {
		return
	}
	r.end = (r.end - 1 + r.size) % r.size
	r.len -= 1
}

func (r *RollingArray[T]) GetLength() int {
	return r.len
}
