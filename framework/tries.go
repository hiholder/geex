package framework

import (
	"fmt"
	"strings"
)

type Tree struct {
	root *node
}

type node struct {
	pattern  string  // 待匹配路由
	part     string  // 路由中的一部分
	children []*node // 子节点
	isWild   bool    // 是否精准匹配
	isLast   bool    // 是否是最后一个
	handler  HandlerFunc // 将handler放到节点上
}

func newNode() *node {
	return &node{
		isLast:  false,
		part: "",
		children: make([]*node, 0),
	}
}

func NewTree() *Tree {
	root := newNode()
	return &Tree{root}
}

// 找的第一个匹配的节点
func (n *node) matchChild(part string) *node {
	if len(n.children) == 0 {
		return nil
	}
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 找到所有匹配的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	if len(n.children) == 0 {
		return nil
	}
	// 获取下一层符合的节点
	nodeList := make([]*node, 0, len(n.children))
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodeList = append(nodeList, child)
		}
	}
	return nodeList
}

func (n *node) insert(pattern string, parts []string, height int) {
	// trie树的高度等于匹配路由长度认为已经匹配完成
	if height == len(parts) {
		// 可以匹配的节点才会有pattern
		n.pattern = pattern
		return
	}
	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
	return
}

// 根据part数组查找对应的节点
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

func (tree *Tree) AddRouter(path string, handler HandlerFunc) error {
	n := tree.root
	if path[0] != '/' {
		return fmt.Errorf("invalid path=%v", path)
	}
	parts := strings.Split(path[1:], "/")
	for i, part := range parts {
		var next *node
		next = n.matchChild(part)
		// 如果没有找到
		if next == nil {
			next = &node{
				part: part,
				children: make([]*node, 0),
			}
			if len(part) > 0 {
				next.isWild = part[0] == ':' || part[0] == '*'
			}
			n.children = append(n.children, next)
			if i == len(parts) - 1 || (len(part) > 0 && part[0] == '*')  {
				if i != len(parts) - 1 {
					return fmt.Errorf("invalid * position, path=%v", path)
				}
				next.isLast = true
				next.handler = handler
				next.pattern = path
				break
			}
		}
		n = next
	}
	return nil
}

func (n *node) matchNode(path string) (*node, error) {
	parts := strings.SplitN(path, "/", 2)
	part := parts[0]
	if strings.HasPrefix(n.part, "*") {
		return n, nil
	}
	children := n.matchChildren(part)
	if children == nil {
		return nil, nil
	}
	// 已经是最后一个节点
	if len(parts) == 1 {
		for _, child := range children {
			if child.isLast {
				return child, nil
			}
		}
		return nil, nil
	}
	for _, child := range children {
		if nodeMatch, err := child.matchNode(parts[1]); nodeMatch != nil || err != nil {
			return nodeMatch, err
		}
	}
	return nil, nil
}

func (tree *Tree) SearchRouter(path string) (HandlerFunc, map[string]string)  {
	if path[0] != '/' {
		return nil, nil //, fmt.Errorf("invalid path=%v", path)
	}
	searchParts := parsePattern(path)
	node, err := tree.root.matchNode(path[1:])
	if err != nil {
		return nil, nil
	}
	if node == nil {
		return nil, nil
	}
	params := make(map[string]string) // 用于解析通配符
	parts := parsePattern(node.pattern)
	for index, part := range parts {
		// 去除路由中的通配符，用来提取参数，比如":lang"输入路径为"username"时，需要在map中保存key为lang，value为username"
		// 去除路由中的":"
		if part[0] == ':' {
			params[part[1:]] = searchParts[index]
		}
		// 去除路由中的"*"
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(searchParts[index:], "/")
			break
		}
	}
	return node.handler, params
}

