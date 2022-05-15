package swarm

import (
	"golang.org/x/net/html"
)

// NodeFilter functions
type NodeFilter func(node *html.Node) bool

func (f1 NodeFilter) And(f2 NodeFilter) NodeFilter {
	return func(node *html.Node) bool {
		return f1(node) && f2(node)
	}
}

func (f1 NodeFilter) Or(f2 NodeFilter) NodeFilter {
	return func(node *html.Node) bool {
		return f1(node) || f2(node)
	}
}

var None NodeFilter = func(node *html.Node) bool {
	return false
}

var All NodeFilter = func(node *html.Node) bool {
	return true
}

func HasAttr(attr string) NodeFilter {
	return func(node *html.Node) bool {
		for _, nAttr := range node.Attr {
			if nAttr.Key == attr {
				return true
			}
		}
		return false
	}
}

func HasAttrs(includedAttrs ...string) NodeFilter {
	filter := None
	for _, inclusion := range includedAttrs {
		filter = filter.Or(HasAttr(inclusion))
	}
	return filter
}

func IsLeafNode(node *html.Node) bool {
	return node.FirstChild == nil
}
