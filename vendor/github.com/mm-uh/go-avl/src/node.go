package avl

type Node struct {
	leftChild  *Node
	rigthChild *Node
	height     int
	key        Key
	treeSize   int
	Value      interface{}
}

type Key interface {
	Less(other interface{}) (bool, error)
}

func NewNode(key Key, value interface{}) *Node {
	return &Node{
		leftChild:  nil,
		rigthChild: nil,
		height:     1,
		key:        key,
		Value:      value,
		treeSize:   1,
	}
}

func (n *Node) isLeaf() bool {
	return n.rigthChild != nil && n.leftChild != nil
}

func (n *Node) hasRChild() bool {
	return n.rigthChild != nil
}

func (n *Node) hasLChild() bool {
	return n.leftChild != nil
}

func getHeight(n *Node) int {
	if n == nil {
		return 0
	}
	return n.height
}

func getSize(n *Node) int {
	if n == nil {
		return 0
	}
	return n.treeSize
}
