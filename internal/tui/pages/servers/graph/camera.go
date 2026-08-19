package graph

import "math"

const (
	minimumZoom = 0.25
	maximumZoom = 4.0
	zoomStep    = 1.15
	fitPadding  = 0.08
)

// camera maps stable world positions onto the current pixel viewport.
type camera struct {
	center        point
	width, height float64
	zoom          float64
	manual        bool
}

func (c camera) worldToScreen(position point) point {
	return point{
		x: (position.x-c.center.x)*c.zoom + c.width/2,
		y: (position.y-c.center.y)*c.zoom + c.height/2,
	}
}

func (c camera) screenToWorld(position point) point {
	zoom := max(c.zoom, minimumZoom)
	return point{
		x: c.center.x + (position.x-c.width/2)/zoom,
		y: c.center.y + (position.y-c.height/2)/zoom,
	}
}

func (c *camera) resize(width, height int) {
	c.width = float64(max(0, width))
	c.height = float64(max(0, height))
	if c.zoom == 0 {
		c.zoom = 1
	}
}

func (c *camera) fit(nodes []node) {
	if len(nodes) == 0 || c.width == 0 || c.height == 0 {
		return
	}
	minimum, maximum := nodes[0].position, nodes[0].position
	for _, node := range nodes[1:] {
		minimum.x = min(minimum.x, node.position.x)
		minimum.y = min(minimum.y, node.position.y)
		maximum.x = max(maximum.x, node.position.x)
		maximum.y = max(maximum.y, node.position.y)
	}
	margin := baseNodeRadius * 3
	minimum.x -= margin
	minimum.y -= margin
	maximum.x += margin
	maximum.y += margin
	c.center = point{x: (minimum.x + maximum.x) / 2, y: (minimum.y + maximum.y) / 2}
	availableWidth := c.width * (1 - 2*fitPadding)
	availableHeight := c.height * (1 - 2*fitPadding)
	c.zoom = min(availableWidth/max(1, maximum.x-minimum.x), availableHeight/max(1, maximum.y-minimum.y))
	c.zoom = clamp(c.zoom, minimumZoom, maximumZoom)
}

func (c *camera) zoomAt(screen point, factor float64) {
	world := c.screenToWorld(screen)
	c.zoom = clamp(c.zoom*factor, minimumZoom, maximumZoom)
	c.center = point{
		x: world.x - (screen.x-c.width/2)/c.zoom,
		y: world.y - (screen.y-c.height/2)/c.zoom,
	}
	c.manual = true
}

func (c *camera) pan(delta point) {
	if math.Abs(delta.x)+math.Abs(delta.y) == 0 {
		return
	}
	c.center.x -= delta.x / c.zoom
	c.center.y -= delta.y / c.zoom
	c.manual = true
}
