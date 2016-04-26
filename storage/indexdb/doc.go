// Copyright 2015 Eleme Inc. All rights reserved.

/*

Package indexdb handles the storage for indexes.

Design

The db file name is index, and the key-value design in leveldb is:

	|--- Key --|------------------ Value (24) -------------------|
	+----------+-----------+-----------+-----------+-------------+
	| Name (X) |  Link (4) | Stamp (4) | Score (8) | Average (8) |
	+----------+-----------+-----------+-----------+-------------+

Cache

To access indexes faster in webapp, indexes are cached in memory, in
a trie with goroutine safety.

Read operations are in cache.

Write operations are to persistence and cache.

*/
package indexdb
