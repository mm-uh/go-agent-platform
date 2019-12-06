package avl

import (
	"errors"
	"fmt"
)

type testNumber struct {
	num uint64
}

func (n testNumber) String() string {
	return fmt.Sprintf("%d", n.num)
}

func (n testNumber) Less(other interface{}) (bool, error) {
	key, ok := other.(testNumber)
	if !ok {
		return false, errors.New("Wrong type")
	}
	return n.num < key.num, nil
}

func newNumber(val uint64) testNumber {
	return testNumber{
		num: val,
	}
}
func InsertTest() bool {
	tree := NewNode(newNumber(10), struct{}{})
	tree = Insert(tree, NewNode(newNumber(20), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(30), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(40), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(50), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(25), struct{}{}))
	return strPreOrder(tree) == "30 20 10 25 40 50 "

}

func DeleteTest() bool {
	tree := NewNode(newNumber(9), struct{}{})
	tree = Insert(tree, NewNode(newNumber(5), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(10), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(1), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(6), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(11), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(0), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(2), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(3), struct{}{}))

	buildOk := strPreOrder(tree) == "9 2 1 0 5 3 6 10 11 "
	//fmt.Println(strPreOrder(tree))
	tree = Delete(tree, newNumber(10))

	deleteOk := strPreOrder(tree) == "2 1 0 9 5 3 6 11 "

	return buildOk && deleteOk
}

func GetKMinsTest() bool {
	tree := NewNode(newNumber(10), struct{}{})
	tree = Insert(tree, NewNode(newNumber(20), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(30), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(40), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(50), struct{}{}))
	tree = Insert(tree, NewNode(newNumber(25), struct{}{}))

	return "10 20" == strFromVector(tree.GetKMins(2)) && "10 20 25" == strFromVector(tree.GetKMins(3)) && "10 20 25 30" == strFromVector(tree.GetKMins(4))

}
