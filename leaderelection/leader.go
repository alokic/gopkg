package leaderelection

/*
A valid leader election algorithm must meet the following conditions:
  1. Termination: the algorithm should finish within a finite time once the leader is selected. In randomized approaches this condition is sometimes weakened (for example, requiring termination with probability 1).
  2. Uniqueness: there is exactly one processor that considers itself as leader.
  3. Agreement: all other processors know who the leader is.
*/

//Leader interface.
type Leader interface {
	IsLeader() (bool, error)
	Stop()
}
