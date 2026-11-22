package widget

// scrollRest keeps what of a scroll is too small to move by, as a precision
// touchpad gives a fraction of a notch or of a pixel, until more adds to it.
type scrollRest float64

// take adds v and returns the whole steps to move by, keeping the rest.
func (r *scrollRest) take(v float64) int {
	*r += scrollRest(v)
	n := int(*r)
	*r -= scrollRest(n)
	return n
}
