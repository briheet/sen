package graph

import "math"

const (
	initialAlpha          = 1.0
	minimumAlpha          = 0.005
	alphaDecay            = 0.03
	dragAlpha             = 0.2
	velocityRetention     = 0.65
	dragVelocityRetention = 0.84
	linkStrength          = 0.08
	dragLinkStrength      = 0.055
	repulsionStrength     = 900.0
	centerStrength        = 0.002
	collisionPadding      = 4.0
	settleVelocity        = 0.05
	maximumLayoutTicks    = 180
	barnesHutThreshold    = 256
	barnesHutTheta        = 0.9
	minimumDistance       = 8.0
)

type simulation struct {
	forces      []point
	tree        []quad
	alpha       float64
	alphaTarget float64
	elastic     bool
}

type quad struct {
	children     [4]int
	center       point
	centerOfMass point
	half, mass   float64
	node         int
}

func newSimulation(kind Kind, nodes []node, edges []edgeModel, root int) simulation {
	configuration := layoutFor(kind)
	for index := range edges {
		edges[index].distance = configuration.linkDistance
	}
	seedNodes(nodes, root, configuration.linkDistance)
	return simulation{alpha: initialAlpha}
}

type layoutConfig struct{ linkDistance float64 }

func layoutFor(kind Kind) layoutConfig {
	if kind == FileGraph {
		return layoutConfig{linkDistance: 96}
	}
	return layoutConfig{linkDistance: 104}
}

func seedNodes(nodes []node, root int, spacing float64) {
	goldenAngle := math.Pi * (3 - math.Sqrt(5))
	for index := range nodes {
		order := index
		if index == root {
			order = 0
		} else if root >= 0 && index < root {
			order++
		}
		radius := spacing * 0.55 * math.Sqrt(float64(order))
		angle := float64(order) * goldenAngle
		nodes[index].position = point{x: radius * math.Cos(angle), y: radius * math.Sin(angle)}
	}
}

func (s *simulation) settle(nodes []node, links []edgeModel) {
	for range maximumLayoutTicks {
		if s.step(nodes, links) {
			break
		}
	}
	s.alpha = 0
	s.alphaTarget = 0
	for index := range nodes {
		nodes[index].velocity = point{}
	}
}

func (s *simulation) reheat() {
	s.alpha = max(s.alpha, dragAlpha)
	s.alphaTarget = dragAlpha
	s.elastic = true
}

func (s *simulation) cool() {
	s.alphaTarget = 0
	s.elastic = false
}

// step advances one deterministic force-layout tick and reports settlement.
func (s *simulation) step(nodes []node, edges []edgeModel) bool {
	if len(nodes) == 0 {
		return true
	}
	s.alpha += (s.alphaTarget - s.alpha) * alphaDecay
	if cap(s.forces) < len(nodes) {
		s.forces = make([]point, len(nodes))
	} else {
		s.forces = s.forces[:len(nodes)]
		clear(s.forces)
	}
	if len(nodes) > barnesHutThreshold {
		s.repulsionBarnesHut(nodes)
	} else {
		s.repulsionExact(nodes)
	}
	spring, retention := linkStrength, velocityRetention
	if s.elastic {
		spring, retention = dragLinkStrength, dragVelocityRetention
	}
	for _, edge := range edges {
		if edge.from == edge.to {
			continue
		}
		from, to := nodes[edge.from].position, nodes[edge.to].position
		delta := point{x: to.x - from.x, y: to.y - from.y}
		length := max(distance(from, to), 0.001)
		magnitude := 2 * spring * (length - edge.distance)
		force := point{x: magnitude * delta.x / length, y: magnitude * delta.y / length}
		// A hub receives a smaller share of each link force so hundreds of
		// adjacent springs cannot launch it—and then the whole layout—away.
		degreeTotal := float64(max(1, nodes[edge.from].degree) + max(1, nodes[edge.to].degree))
		fromShare := float64(max(1, nodes[edge.to].degree)) / degreeTotal
		toShare := float64(max(1, nodes[edge.from].degree)) / degreeTotal
		s.forces[edge.from].x += force.x * fromShare
		s.forces[edge.from].y += force.y * fromShare
		s.forces[edge.to].x -= force.x * toShare
		s.forces[edge.to].y -= force.y * toShare
	}
	for index := range nodes {
		node := &nodes[index]
		s.forces[index].x -= node.position.x * centerStrength
		s.forces[index].y -= node.position.y * centerStrength
		if node.fixed {
			node.velocity = point{}
			continue
		}
		node.velocity.x = (node.velocity.x + s.forces[index].x*s.alpha) * retention
		node.velocity.y = (node.velocity.y + s.forces[index].y*s.alpha) * retention
		node.position.x += node.velocity.x
		node.position.y += node.velocity.y
	}
	return s.settled(nodes)
}

func (s *simulation) settled(nodes []node) bool {
	if s.alpha > minimumAlpha {
		return false
	}
	for _, node := range nodes {
		if math.Hypot(node.velocity.x, node.velocity.y) > settleVelocity {
			return false
		}
	}
	return true
}

func nodeMass(node node) float64 { return math.Sqrt(float64(min(node.degree, 8) + 1)) }

func (s *simulation) repulsionExact(nodes []node) {
	for first := range nodes {
		for second := first + 1; second < len(nodes); second++ {
			delta := point{x: nodes[second].position.x - nodes[first].position.x, y: nodes[second].position.y - nodes[first].position.y}
			if delta == (point{}) {
				delta = deterministicJitter(nodes[first].id, nodes[second].id)
			}
			distanceSquared := max(delta.x*delta.x+delta.y*delta.y, minimumDistance*minimumDistance)
			length := math.Sqrt(distanceSquared)
			magnitude := repulsionStrength * nodeMass(nodes[first]) * nodeMass(nodes[second]) / distanceSquared
			minimum := baseNodeRadius*(nodes[first].scale+nodes[second].scale) + collisionPadding
			if length < minimum {
				magnitude += (minimum - length) * 0.8
			}
			force := point{x: magnitude * delta.x / length, y: magnitude * delta.y / length}
			s.forces[first].x -= force.x
			s.forces[first].y -= force.y
			s.forces[second].x += force.x
			s.forces[second].y += force.y
		}
	}
}

func deterministicJitter(first, second uint64) point {
	value := first*0x9e3779b97f4a7c15 ^ second
	angle := float64(value%6283) / 1000
	return point{x: math.Cos(angle), y: math.Sin(angle)}
}

func (s *simulation) repulsionBarnesHut(nodes []node) {
	s.buildTree(nodes)
	for index := range nodes {
		s.applyTreeForce(nodes, index, 0)
	}
}

func (s *simulation) buildTree(nodes []node) {
	minimum, maximum := nodes[0].position, nodes[0].position
	for _, node := range nodes[1:] {
		minimum.x, minimum.y = min(minimum.x, node.position.x), min(minimum.y, node.position.y)
		maximum.x, maximum.y = max(maximum.x, node.position.x), max(maximum.y, node.position.y)
	}
	center := point{x: (minimum.x + maximum.x) / 2, y: (minimum.y + maximum.y) / 2}
	half := max(maximum.x-minimum.x, maximum.y-minimum.y)/2 + 1
	s.tree = s.tree[:0]
	s.tree = append(s.tree, newQuad(center, half))
	for index := range nodes {
		s.insertNode(nodes, 0, index, 0)
	}
}

func newQuad(center point, half float64) quad {
	return quad{center: center, half: half, node: -1, children: [4]int{-1, -1, -1, -1}}
}

func (s *simulation) insertNode(nodes []node, cellIndex, nodeIndex, depth int) {
	cell := &s.tree[cellIndex]
	mass := nodeMass(nodes[nodeIndex])
	total := cell.mass + mass
	cell.centerOfMass.x = (cell.centerOfMass.x*cell.mass + nodes[nodeIndex].position.x*mass) / total
	cell.centerOfMass.y = (cell.centerOfMass.y*cell.mass + nodes[nodeIndex].position.y*mass) / total
	cell.mass = total
	if cell.node == -1 && cell.children[0] == -1 {
		cell.node = nodeIndex
		return
	}
	if depth >= 24 {
		return
	}
	if cell.children[0] == -1 {
		existing := cell.node
		cell.node = -1
		s.splitQuad(cellIndex)
		s.insertNode(nodes, s.childFor(cellIndex, nodes[existing].position), existing, depth+1)
	}
	s.insertNode(nodes, s.childFor(cellIndex, nodes[nodeIndex].position), nodeIndex, depth+1)
}

func (s *simulation) splitQuad(index int) {
	center, half := s.tree[index].center, s.tree[index].half/2
	children := [4]int{}
	for child := range 4 {
		x := center.x + math.Copysign(half, float64(child&1)-0.5)
		y := center.y + math.Copysign(half, float64(child&2)-1)
		children[child] = len(s.tree)
		s.tree = append(s.tree, newQuad(point{x: x, y: y}, half))
	}
	s.tree[index].children = children
}

func (s *simulation) childFor(index int, position point) int {
	child := 0
	if position.x >= s.tree[index].center.x {
		child |= 1
	}
	if position.y >= s.tree[index].center.y {
		child |= 2
	}
	return s.tree[index].children[child]
}

func (s *simulation) applyTreeForce(nodes []node, nodeIndex, cellIndex int) {
	cell := s.tree[cellIndex]
	if cell.mass == 0 || (cell.children[0] == -1 && cell.node == nodeIndex) {
		return
	}
	position := nodes[nodeIndex].position
	delta := point{x: cell.centerOfMass.x - position.x, y: cell.centerOfMass.y - position.y}
	distanceSquared := max(delta.x*delta.x+delta.y*delta.y, minimumDistance*minimumDistance)
	length := math.Sqrt(distanceSquared)
	contains := math.Abs(position.x-cell.center.x) <= cell.half && math.Abs(position.y-cell.center.y) <= cell.half
	if cell.children[0] == -1 || (!contains && cell.half*2/length < barnesHutTheta) {
		magnitude := repulsionStrength * nodeMass(nodes[nodeIndex]) * cell.mass / distanceSquared
		if cell.children[0] == -1 && cell.node >= 0 {
			minimum := baseNodeRadius*(nodes[nodeIndex].scale+nodes[cell.node].scale) + collisionPadding
			if length < minimum {
				magnitude += (minimum - length) * 0.8
			}
		}
		s.forces[nodeIndex].x -= magnitude * delta.x / length
		s.forces[nodeIndex].y -= magnitude * delta.y / length
		return
	}
	for _, child := range cell.children {
		s.applyTreeForce(nodes, nodeIndex, child)
	}
}
