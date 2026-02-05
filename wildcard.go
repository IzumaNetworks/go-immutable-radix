package iradix

import "bytes"

// MatchWithWildcards checks if a key matches any pattern in the tree, considering wildcard
// patterns at dot-separated segment boundaries. This performs a single tree traversal,
// checking for wildcard matches during the descent through the tree.
//
// For example, given key "tenant.abc123.project.xyz789.member.add", it checks for:
//   - "*" (universal wildcard)
//   - "tenant.*"
//   - "tenant.abc123.*"
//   - "tenant.abc123.project.*"
//   - "tenant.abc123.project.xyz789.*"
//   - "tenant.abc123.project.xyz789.member.*"
//   - "tenant.abc123.project.xyz789.member.add" (exact match)
//
// The function returns true if any match is found.
func (n *Node[T]) MatchWithWildcards(key []byte) bool {
	if len(key) == 0 {
		_, ok := n.Get(key)
		return ok
	}

	// Check for universal wildcard "*"
	if _, ok := n.Get([]byte("*")); ok {
		return true
	}

	// Perform tree traversal while checking for wildcards at dot boundaries
	return n.matchWithWildcardsFrom(key, key)
}

// matchWithWildcardsFrom performs the traversal.
// - originalKey: the full original key (never changes)
// - search: the remaining part to search
func (n *Node[T]) matchWithWildcardsFrom(originalKey, search []byte) bool {
	// Base case: search exhausted - check for exact match
	if len(search) == 0 {
		return n.isLeaf()
	}

	// Before looking for the specific edge, check if there's a wildcard '*' edge
	// This handles patterns like "tenant.*" where '*' is a direct child
	_, wildcardNode := n.getEdge('*')
	if wildcardNode != nil && wildcardNode.isLeaf() {
		// Found a wildcard pattern at this level
		return true
	}

	// Look for the next edge matching our search
	_, next := n.getEdge(search[0])
	if next == nil {
		return false
	}

	// Check if this node's prefix represents a wildcard pattern (ends with ".*")
	if len(next.prefix) >= 2 &&
	   next.prefix[len(next.prefix)-2] == '.' &&
	   next.prefix[len(next.prefix)-1] == '*' &&
	   next.isLeaf() {
		// This is a wildcard pattern. Check if search matches the prefix before ".*"
		wildcardPrefix := next.prefix[:len(next.prefix)-2] // Remove ".*"
		if bytes.HasPrefix(search, wildcardPrefix) {
			// The search key matches this wildcard pattern!
			// Check if there's a dot after the prefix (or it's the end)
			if len(search) == len(wildcardPrefix) ||
			   (len(search) > len(wildcardPrefix) && search[len(wildcardPrefix)] == '.') {
				return true
			}
		}
	}

	// Check if the node's prefix matches our search for normal traversal
	if bytes.HasPrefix(search, next.prefix) {
		// Prefix matches - consume it and recurse
		return next.matchWithWildcardsFrom(originalKey, search[len(next.prefix):])
	} else if bytes.HasPrefix(next.prefix, search) {
		// Search is shorter than prefix but matches what we have
		// Check if this is an exact match (we've consumed all of search)
		return next.isLeaf() && len(search) == len(next.prefix)
	}

	// Prefix mismatch
	return false
}
