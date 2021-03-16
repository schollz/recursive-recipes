// Directed Acyclic Graph implementation in golang.
package dag

import (
	"fmt"
)

type Dag struct {
	nodes map[string]*Node
}

type Node struct {
	name string
	val  interface{}

	indegree int
	children []*Node
}

func New() *Dag {
	this := new(Dag)
	this.nodes = make(map[string]*Node)
	return this
}

func (this *Dag) AddVertex(name string, val interface{}) *Node {
	node := &Node{name: name, val: val}
	this.nodes[name] = node
	return node
}

func (this *Dag) AddEdge(from, to string) {
	fromNode := this.nodes[from]
	toNode := this.nodes[to]
	fromNode.children = append(fromNode.children, toNode)
	toNode.indegree++
}

// func (this *Dag) MakeDotGraph(fn string) string {
// 	file, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	sb := str.NewStringBuilder()
// 	sb.WriteString("digraph depgraph {\n\trankdir=LR;\n")
// 	for _, node := range this.nodes {
// 		node.dotGraph(sb)
// 	}
// 	sb.WriteString("}\n")
// 	file.WriteString(sb.String())
// 	return sb.String()
// }

func (this *Dag) HasPathTo(that string) bool {
	return false
}

func (this *Node) dotGraph(sb *str.StringBuilder) {
	if len(this.children) == 0 {
		sb.WriteString(fmt.Sprintf("\t\"%s\";\n", this.name))
		return
	}

	for _, child := range this.children {
		sb.WriteString(fmt.Sprintf(`%s -> %s [label="%v"]`, this.name, child.name, this.val))
		sb.WriteString("\r\n")
	}
}

func (this *Node) Children() []*Node {
	return this.children
}
