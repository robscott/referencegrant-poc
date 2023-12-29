package main

func main() {
	NewController()

	// ReferenceGrant changes
	// - New: Generate new RoleBindings for pattern
	// - Update: Change/diff RoleBindings for pattern
	// - Delete: Update/remove RoleBindings for pattern

	// ClusterReferenceConsumer changes
	// - New: Add subject to all role bindings for pattern
	// - Update: Change subject == change to subject in relevant role bindings
	// - Delete Remove subject from all role bindings

	// ClusterReferencePattern changes
	// - New: Add role bindings for all patterns with empty subject, use predefined label, r+w lock cache for pattern until it's built out
	// - Update: Change role bindings for all patterns, r+w lock cache for pattern until it's built out
	// - Delete: Delete role bindings, r+w lock cache for pattern until it's deleted

	// ClusterReferencePattern Resource changes
	// - New: Create RoleBinding for pattern
	//
	//

	// ReconcilePattern
	// 1) Get all consumers of pattern, derive subjects from that
	// 2) Get current set of role bindings generated for this pattern via label
	// 3) Get desired set of role bindings from cache of pattern references - needs to be rebuilt for some ClusterReferencePattern changes - those should lock cache
	// 4) Update existing role bindings, create missing ones, delete unnecessary
}
