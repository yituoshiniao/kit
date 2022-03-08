package xslice

import "gitlab.intsig.net/cs-server2/kit/xtype"

func Paged(list []interface{}, cursor int64, size int64) (outList []interface{}, outPager xtype.Pager) {
	nextCursor := cursor + int64(size)
	total := int64(len(list))

	var hasMore bool
	if total > nextCursor {
		hasMore = true
		outList = list[cursor:nextCursor]
	} else {
		outList = list[cursor:]
	}

	outPager = xtype.NewPager(nextCursor, size, hasMore)

	return outList, outPager
}

func PagedInt64(list []int64, cursor int64, size int64) (outList []int64, outPager xtype.Pager) {
	nextCursor := cursor + int64(size)
	total := int64(len(list))
	if cursor > total {
		return outList, xtype.NewPager(cursor, size, false)
	}

	var hasMore bool
	if total > nextCursor {
		hasMore = true
		outList = list[cursor:nextCursor]
	} else {
		outList = list[cursor:]
	}

	outPager = xtype.NewPager(nextCursor, size, hasMore)
	return outList, outPager
}

func CursorPaged(list []int64, cursor int64, size int32) (outList []int64, outPager xtype.CursorPager) {
	nextCursor := cursor + int64(size)
	total := int64(len(list))
	total2 := int32(len(list))

	var hasMore bool
	if total > nextCursor {
		hasMore = true
		outList = list[cursor:nextCursor]
	} else {
		outList = list[cursor:]
	}

	outPager = xtype.NewCursorPager(nextCursor, size, hasMore, total2)

	return outList, outPager
}
