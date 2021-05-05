package dark

import (
	"strings"
)

/*
前缀树结构来实现路由查询
*/
type node struct {
	pattern  string  // 待匹配路由
	part     string  // 路由中的部分内容
	children []*node // 子节点，
	isWild   bool    // 是否模糊匹配，提供两种路由参数的通配符 ：和 *
}

// 第一个匹配成功的节点用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的子节点
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// 只有叶子节点的pattern才会被设置，路径上的pattern值为空
// 如果匹配规则中的一部分part没有在前缀树中则添加节点
// 当匹配规则的所有part都已经匹配完成递归终止
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// 递归终止条件，parts所有匹配都已匹配完成或者当前节点的part为通用匹配 "*"
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") || strings.HasPrefix(n.part, ":") {
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
