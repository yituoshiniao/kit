//go:generate msgp -tests=false
package xtype

// Deprecated
// Pager 游标分页,已废弃，请使用CursorPager.
type Pager struct {
	Cursor  int64 `json:"cursor"`
	Size    int64 `json:"size"`
	HasMore int32 `json:"hasMore"`
}

// Deprecated
// Pager 游标分页,已废弃，请使用CursorPager.
func NewPager(nextCursor int64, perPageSize int64, hasMore bool) Pager {
	var more int32
	if hasMore {
		more = 1
	}

	return Pager{
		Cursor:  nextCursor,
		Size:    perPageSize,
		HasMore: more,
	}
}

// CursorPager 游标分页.
// 应用场景:一般用于APP,Web游标分页(下一页)
type CursorPager struct {
	Cursor  int64 `json:"cursor"`
	Size    int32 `json:"size"`
	HasMore bool  `json:"hasMore"`
	Total   int32 `json:"total,omitempty"`
}

func NewCursorPager(nextCursor int64, perPageSize int32, hasMore bool, total int32) CursorPager {
	return CursorPager{
		Cursor:  nextCursor,
		Size:    perPageSize,
		HasMore: hasMore,
		Total:   total,
	}
}

// PagePager 页码形式分页.
// 应用场景:一般用于后台页码分页
type PagePager struct {
	// 当前页码
	Page int64 `json:"page"`
	// 每页记录条数
	Size int32 `json:"size"`
	// 总数
	Total int32 `json:"total"`
}

func NewPagePager(page int64, size int32, total int32) PagePager {
	return PagePager{Page: page, Size: size, Total: total}
}
