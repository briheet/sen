package graph

// edgeModel is one analyzed function call.
type edgeModel struct {
	from     int
	to       int
	distance float64
	active   bool
}
