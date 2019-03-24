package queue

type Queue interface {
	Push(e interface{})
	Pop() interface{}
	Top() interface{}
	Empty() bool
	Size() int
}

//Not thread safe
type queue struct {
	q []interface{}
}

func New() Queue {
	q := new(queue)
	q.q = []interface{}{}
	return q
}

func (q *queue) Push(e interface{}) {
	q.q = append(q.q, e)
}

func (q *queue) Pop() interface{} {
	top := q.q[0]
	q.q = q.q[1:]
	return top
}

func (q *queue) Top() interface{} {
	return q.q[0]
}

func (q *queue) Empty() bool {
	return len(q.q) == 0
}

func (q *queue) Size() int {
	return len(q.q)
}
