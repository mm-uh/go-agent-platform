package avl

import "fmt"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func strPreOrder(tree *Node) string {
	if tree == nil {
		return ""
	}
	return fmt.Sprintf("%s %s%s", tree.key, strPreOrder(tree.leftChild), strPreOrder(tree.rigthChild))

}

func strFromVector(list []*Node) string {
	result := ""
	for _, node := range list {
		result = fmt.Sprintf("%s %s", result, node.key)
	}
	return result[1:]
}
