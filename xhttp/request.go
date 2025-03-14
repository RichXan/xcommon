package xhttp

type PageReq struct {
	Current int    `json:"current" form:"current" binding:"omitempty,min=1" default:"1"` // 当前页
	Size    int    `json:"size" form:"size" binding:"omitempty,min=1" default:"10"`      // 每页大小
	Order   string `json:"order" form:"order" binding:"omitempty" default:"id"`          // 排序字段
}

func (p PageReq) GetCurrent() int {
	if p.Current == 0 {
		return 1
	}
	return p.Current
}

func (p PageReq) GetSize() int {
	if p.Size <= 0 {
		return 10
	}
	return p.Size
}

func (p PageReq) GetOrder() string {
	if p.Order == "" {
		return "id desc"
	}
	return p.Order
}

func (p PageReq) GetOffset() int {
	if p.Current <= 0 {
		return 0
	}
	return (p.GetCurrent() - 1) * p.GetSize()
}

func (p PageReq) GetLimit() int {
	return p.GetSize()
}
