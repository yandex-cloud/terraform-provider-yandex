package multichoose

type MultiChoose struct {
	selected []uint64
	length   int
	limit    int
}

func (mc *MultiChoose) getSelected(index int) bool {
	if index < 0 || index > mc.length {
		return false
	}
	i := uint64(index) >> 6
	j := uint64(index) & 0x3F
	return (mc.selected[i] & (1 << j)) != 0
}

func (mc *MultiChoose) setSelected(index int, value bool) {
	if index < 0 || index > mc.length {
		return
	}
	i := uint64(index) >> 6
	j := uint64(index) & 0x3F
	if value {
		mc.selected[i] |= (1 << j)
	} else {
		mc.selected[i] &^= (1 << j)
	}
}

func New(length int) *MultiChoose {
	return &MultiChoose{
		selected: make([]uint64, (length>>6)+1),
		length:   length,
		limit:    -1,
	}
}

func (mc MultiChoose) Count() int {
	count := 0
	for i := 0; i <= mc.length; i++ {
		if mc.getSelected(i) {
			count++
		}
	}
	return count
}

func (mc *MultiChoose) Select(index int) {
	if mc.limit >= 0 && mc.Count() >= mc.limit {
		return
	}

	mc.setSelected(index, true)
}

func (mc *MultiChoose) Deselect(index int) {
	mc.setSelected(index, false)
}

func (mc MultiChoose) IsSelected(index int) bool {
	return mc.getSelected(index)
}
