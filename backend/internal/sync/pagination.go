package sync

// paginationMeta derives has_more and the page length when fetching pageSize+1 rows.
func paginationMeta(fetched, pageSize int) (hasMore bool, count int) {
	hasMore = fetched > pageSize
	if hasMore {
		count = pageSize
	} else {
		count = fetched
	}
	return hasMore, count
}
