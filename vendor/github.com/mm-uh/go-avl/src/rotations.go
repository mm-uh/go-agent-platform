package avl

func (n *Node) rRotation() *Node {
	x := n.leftChild
	t := x.rigthChild

	n.leftChild = t
	x.rigthChild = n

	n.updateSize()
	n.updateHeight()
	x.updateSize()
	x.updateHeight()

	return x
}

func (n *Node) lRotation() *Node {
	x := n.rigthChild
	t := x.leftChild

	n.rigthChild = t
	x.leftChild = n

	n.updateSize()
	n.updateHeight()
	x.updateSize()
	x.updateHeight()

	return x
}
