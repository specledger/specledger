// Package deps provides artifact path discovery, cache operations, and reference resolution
// for SpecLedger dependencies.
//
// The package handles:
// - Auto-discovery of artifact_path from SpecLedger repositories
// - Manual artifact_path specification for non-SpecLedger repos
// - Cache operations for ~/.specledger/cache/
// - Reference resolution using alias:artifact syntax
package deps
