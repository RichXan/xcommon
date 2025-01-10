package xhttp

type PageReq struct {
	Current int    `json:"current" form:"current"` // 当前页
	Size    int    `json:"size" form:"size"`       // 每页大小
	Order   string `json:"order" form:"order"`     // 排序字段
}

func (p PageReq) GetCurrent() int {
	if p.Current == 0 {
		return 1
	}
	return p.Current
}

func (p PageReq) GetSize() int {
	if p.Size == 0 {
		return 10
	}
	return p.Size
}

func (p PageReq) GetOffset() int {
	return (p.GetCurrent() - 1) * p.GetSize()
}
