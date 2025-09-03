package object

import (
	"github.com/google/btree"
)

type sourceTree struct {
	tree *btree.BTree
}

// Initialize a new source tree for the environment
func initSourceTree() *sourceTree {
	t := &sourceTree{tree: btree.New(2)}

	return t
}
