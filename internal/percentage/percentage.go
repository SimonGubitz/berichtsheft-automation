package percentage

type BoolPercentage struct {
	n map[bool]int
}

func NewBoolPercentage() *BoolPercentage {
	n := make(map[bool]int)
	return &BoolPercentage{
		n: n,
	}
}

func (p *BoolPercentage) Inc(b bool) {
	p.n[b] += 1
}

func (p *BoolPercentage) AddN(n int, b bool) {
	p.n[b] += n
}

func (p *BoolPercentage) Calculate() int {

	if len(p.n) == 0 {
		return 0
	}

	return (p.n[true] / (p.n[true] + p.n[false])) * 100
}
