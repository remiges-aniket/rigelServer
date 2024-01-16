package etcdls

import "strings"

type Node struct {
	Name     string
	Children []*Node
	IsLeaf   bool
	Value    string
}

func NewNode(name string, children []*Node, isLeaf bool, value string) *Node {

	tempN := &Node{
		Name:     name,
		Children: children,
		IsLeaf:   isLeaf,
		Value:    value,
	}
	return tempN
}

// func (n *Node) getChildren(prefix string) []*Node {
// 	var matchingChildren []*Node
// 	for _, child := range n.Children {
// 		if strings.HasPrefix(child.Name, prefix) {
// 			matchingChildren = append(matchingChildren, child)
// 		}
// 	}
// 	return matchingChildren
// }

func BuildTree(te []*Node, data map[string]string) ([]*Node, map[string]string) {
	var remainingData map[string]string
	if len(data) == 0 {
		return te, nil
	}

	var newTree []*Node
	for k, v := range data {
		parts := strings.Split(k, "/")
		var subtree []*Node
		subtree, remainingData = BuildTree(nil, make(map[string]string))
		root := NewNode(k, subtree, true, v)
		newTree = append(newTree, root)
		if len(parts) > 1 {
			for _, v := range parts {
				if v != "" && !strings.HasSuffix(k, v) {
					node := NewNode(v, nil, false, "")
					root.Children = append(root.Children, node)
					root = node
				}
			}
		}
	}

	return newTree, remainingData
}

// in main

// func buildTree() {
// []keys := get all the keys from etcd
// for each path in []keys
//	 addPath(path)

// }
