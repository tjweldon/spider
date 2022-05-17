package swarm

import (
	"golang.org/x/net/html"
)

// NodeFilter functions return true if a crawler should crawl the node, false
// if it should move on.
type NodeFilter func(node *html.Node) bool

// And is a higher order function that combines two filters f1, f2 such that
// the resulting filter is f12(node) = f1(node) and f2(node)
func (f1 NodeFilter) And(f2 NodeFilter) NodeFilter {
	return func(node *html.Node) bool {
		return f1(node) && f2(node)
	}
}

// Or is a higher order function that combines two filters f1, f2 such that
// the resulting filter is f12(node) = f1(node) or f2(node)
func (f1 NodeFilter) Or(f2 NodeFilter) NodeFilter {
	return func(node *html.Node) bool {
		return f1(node) || f2(node)
	}
}

// None is the trivial, scrape nothing implementation.
// The most strict filter.
var None NodeFilter = func(node *html.Node) bool {
	return false
}

// All is the trivial, scrape everything implementation.
// The least strict filter.
var All NodeFilter = func(node *html.Node) bool {
	return true
}

// HasAttr is a factory for NodeFilter functions. Given an attribute key,
// HasAttr returns a NodeFilter that scrapes nodes that habe that attribute key.
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

// HasAttrs is the reduction of HasAttr filters defined by each includedAttr
// under Or.
func HasAttrs(includedAttrs ...string) NodeFilter {
	filter := None
	for _, inclusion := range includedAttrs {
		filter = filter.Or(HasAttr(inclusion))
	}
	return filter
}

// IsLeafNode is a filter that returns true if the node in question has no child
// nodes.
func IsLeafNode(node *html.Node) bool {
	return node.FirstChild == nil
}
