package avl

//If existe a node with key == value return true and the node with key == value, if there is not return false and the node that would be it's father
func (n *Node) nodeByKey(value Key) (bool, *Node) {
	valueLess, _ := value.Less(n.key)
	valueBigger, _ := n.key.Less(value)
	valueEqual := (!valueBigger) && (!valueLess)
	if valueEqual {
		return true, n
	}
	if (valueBigger && !n.hasRChild()) || (valueLess && !n.hasLChild()) {
		return false, n
	}
	if valueLess {
		return n.leftChild.nodeByKey(value)
	}
	return n.rigthChild.nodeByKey(value)
}

func (n *Node) HasKey(value Key) bool {
	answ, _ := n.nodeByKey(value)
	return answ
}

func (n *Node) GetNode(value Key) (bool, *Node) {
	exist, node := n.nodeByKey(value)
	if !exist {
		return false, nil
	}
	return true, node
}

func (n *Node) updateHeight() {
	n.height = max(getHeight(n.rigthChild), getHeight(n.leftChild)) + 1
}

func (n *Node) updateSize() {
	n.treeSize = getSize(n.rigthChild) + getSize(n.leftChild) + 1
}

func (n *Node) GetKMins(k int) []*Node {

	if k == 0 {
		return make([]*Node, 0)
	}
	if getSize(n) <= k {
		return inOrder(n)
	}
	if getSize(n.leftChild) >= k {
		return n.leftChild.GetKMins(k)
	}
	answ := inOrder(n.leftChild)
	answ = append(answ, n)
	if len(answ) < k {
		answ = append(answ, n.rigthChild.GetKMins(k-len(answ))...)
	}

	return answ

}

func inOrder(n *Node) []*Node {
	if n == nil {
		return make([]*Node, 0)
	}
	answ := inOrder(n.leftChild)
	answ = append(answ, n)
	return append(answ, inOrder(n.rigthChild)...)
}

func Insert(tree, newNode *Node) *Node {
	if tree == nil {
		return newNode
	}

	newNodeLess, _ := newNode.key.Less(tree.key)
	newNodeBigger, _ := tree.key.Less(newNode.key)
	if newNodeLess {
		tree.leftChild = Insert(tree.leftChild, newNode)
	} else if newNodeBigger {
		tree.rigthChild = Insert(tree.rigthChild, newNode)
	} else {
		return tree
	}
	tree.updateSize()
	tree.updateHeight()
	bl := tree.balanceFactor()

	if bl > 1 {
		newNodeLess, _ = newNode.key.Less(tree.leftChild.key)
		newNodeBigger, _ = tree.leftChild.key.Less(newNode.key)
		if newNodeLess {
			return tree.rRotation()
		}
		if newNodeBigger {
			tree.leftChild = tree.leftChild.lRotation()
			return tree.rRotation()
		}

	}

	if bl < -1 {
		newNodeBigger, _ = tree.rigthChild.key.Less(newNode.key)
		newNodeLess, _ = newNode.key.Less(tree.rigthChild.key)

		if newNodeBigger {
			return tree.lRotation()
		}

		if newNodeLess {
			tree.rigthChild = tree.rigthChild.rRotation()
			return tree.lRotation()
		}

	}

	return tree
}

func Delete(tree *Node, key Key) *Node {
	if tree == nil {
		return tree
	}

	keyLess, _ := key.Less(tree.key)
	keyBigger, _ := tree.key.Less(key)
	if keyLess {
		tree.leftChild = Delete(tree.leftChild, key)
	} else if keyBigger {
		tree.rigthChild = Delete(tree.rigthChild, key)
	} else {
		if tree.leftChild == nil || tree.rigthChild == nil {
			if tree.leftChild != nil {
				return tree.leftChild
			} else if tree.rigthChild != nil {
				return tree.rigthChild
			} else {
				return nil
			}
		}
		newRoot := getMinNode(tree)
		newRoot.leftChild = tree.leftChild
		newRoot.rigthChild = Delete(tree, newRoot.key)
		tree = newRoot
	}

	tree.updateHeight()
	tree.updateSize()

	bl := tree.balanceFactor()

	if bl > 1 && (tree.leftChild.balanceFactor() >= 0) {
		return tree.rRotation()
	}

	if bl < -1 && (tree.rigthChild.balanceFactor() <= 0) {
		return tree.lRotation()
	}

	if bl > 1 && (tree.leftChild.balanceFactor() < 0) {
		tree.leftChild = tree.leftChild.lRotation()
		return tree.rRotation()
	}

	if bl < -1 && (tree.rigthChild.balanceFactor() > 0) {
		tree.rigthChild = tree.rigthChild.rRotation()
		return tree.lRotation()
	}

	return tree

}

func (n *Node) GetSize() int {
	return getSize(n)
}

func (n *Node) balanceFactor() int {
	return getHeight(n.leftChild) - getHeight(n.rigthChild)
}

func getMinNode(tree *Node) *Node {
	for tree != nil && tree.leftChild != nil {
		tree = tree.leftChild
	}
	return tree
}
func getSuccessorNode(tree *Node) *Node {
	return getMinNode(tree.rigthChild)
}

func (n *Node) GetMax() *Node {
	if !n.hasRChild() {
		return n
	}
	return n.rigthChild.GetMax()
}
