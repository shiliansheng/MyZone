package views

import "fmt"

func SetPager(baselink, module string, count, limit, page int) string {
	if count <= limit {
		return ""
	}
	var (
		pageCnt   int    = count / limit
		pageStr   string = ""
		pageLimit int    = 10
		pageBegin int    = 1
		pageEnd   int
	)
	if page == 0 {
		page = 1
	}
	if count%limit != 0 {
		pageCnt++
	}
	if module != "" {
		baselink += "/" + module
	}
	if page > 2 {
		pageBegin = page - 2
	}
	if pageLimit > pageCnt {
		pageLimit = pageCnt
	}
	pageEnd = pageBegin + pageLimit
	if pageEnd > pageCnt {
		pageEnd = pageCnt
		pageBegin = pageCnt - pageLimit
	}
	if pageBegin < 1 {
		pageBegin = 1
	}
	// if page == 1 {
	// 	pageStr += fmt.Sprintf("<a class=\"page-item page-link disabled\">&laquo;</a>\n")
	// } else {
	// 	pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s?page=%d&limit=%d\">&laquo;</a>\n", baselink, page-1, limit)
	// }
	pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s\">首页</a>\n", baselink)
	if page > 1 {
		pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s?page=%d\">PRE</a>\n", baselink, page - 1)
	}
	for i := pageBegin; i <= pageEnd; i++ {
		if i == page {
			pageStr += fmt.Sprintf("<a class=\"page-item page-link active\">%d</a>\n", i)
		} else {
			pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s?page=%d\">%d</a>\n", baselink, i, i)
		}
	}
	if page != pageCnt {
		pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s?page=%d\">NEXT</a>\n", baselink, page + 1)
	}
	pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s?page=%d\">%d/尾页</a>\n", baselink, pageCnt, pageCnt)
	// if page == pageCnt {
	// 	pageStr += fmt.Sprintf("<a class=\"page-item page-link disabled\">&raquo;</a>\n")
	// } else {
	// 	pageStr += fmt.Sprintf("<a class=\"page-item page-link\" href=\"%s?page=%d&limit=%d\">&raquo;</a>\n", baselink, page+1, limit)
	// }
	return pageStr
}
