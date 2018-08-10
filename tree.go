package insidetree

import (
	"container/list"

	"github.com/golang/geo/s2"
)

// Node is the graph unit
type Node struct {
	sub   [4]*Node
	value []interface{}
}

// Tree root stores nodes starting by the 6 faces of the cube
type Tree struct {
	face [6]*Node
}

// NewTree returns a new tree ready for use
func NewTree() *Tree {
	t := &Tree{}
	for i := 0; i < 6; i++ {
		t.face[i] = &Node{}
	}

	return t
}

// Index a cell
func (t *Tree) Index(c s2.CellID, v interface{}) {
	l := c.Level()

	// current node
	cn := t.face[c.Face()]
	// for each level identify the subsquare
	for cl := 1; cl <= l; cl++ {
		pos := c.ChildPosition(cl)

		if cn.sub[pos] == nil {
			cn.sub[pos] = &Node{}
		}
		cn = cn.sub[pos]
	}
	cn.value = append(cn.value, v)
}

// Stab look through the tree for matching sub cells
func (t *Tree) Stab(c s2.CellID) []interface{} {
	l := c.Level()
	m := make(map[interface{}]struct{})

	// current node
	cn := t.face[c.Face()]
	// for each level identify the subsquare
	// going into the last subsqauare to find the value
	for cl := 1; cl <= l+1; cl++ {
		pos := c.ChildPosition(cl)

		if len(cn.value) != 0 {
			for _, v := range cn.value {
				m[v] = struct{}{}
			}
		}
		if cn.sub[pos] == nil {
			break
		}
		cn = cn.sub[pos]
	}

	res := make([]interface{}, len(m))
	i := 0
	for n := range m {
		res[i] = n
		i++
	}
	return res
}

// Mask looks for all value masked by a cell and sub cells
func (t *Tree) Mask(c s2.CellID) []interface{} {
	l := c.Level()
	// current node
	cn := t.face[c.Face()]

	m := make(map[interface{}]struct{})
	visited := make(map[*Node]struct{})

	s := list.New()
	// for each level identify the subsquare
	// going through allsubsqauare to find the value using DFS

	for cl := 1; cl <= l; cl++ {
		pos := c.ChildPosition(cl)

		if cn.sub[pos] == nil {
			return nil
		}
		cn = cn.sub[pos]
	}

	s.PushBack(cn)

	for e := s.Front(); e != nil; e = e.Next() {
		n := e.Value.(*Node)

		if _, exist := visited[n]; exist {
			continue
		}
		visited[n] = struct{}{}
		for _, v := range n.value {
			m[v] = struct{}{}
		}
		for _, sn := range n.SubNodes() {
			s.PushBack(sn)
		}
	}

	res := make([]interface{}, len(m))
	i := 0
	for n := range m {
		res[i] = n
		i++
	}
	return res
}

// SubNodes returns the 4 subnodes if existing
func (n *Node) SubNodes() []*Node {
	var res []*Node
	for _, sn := range n.sub {
		if sn != nil {
			res = append(res, sn)
		}
	}
	return res
}
