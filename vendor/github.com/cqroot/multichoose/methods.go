package multichoose

func (mc MultiChoose) Length() int {
	return mc.length
}

func (mc *MultiChoose) SetLength(length int) {
	mc.length = length
}

func (mc MultiChoose) Limit() int {
	return mc.limit
}

func (mc *MultiChoose) SetLimit(limit int) {
	mc.limit = limit
}

func (mc *MultiChoose) Toggle(index int) {
	if mc.IsSelected(index) {
		mc.Deselect(index)
	} else {
		mc.Select(index)
	}
}
