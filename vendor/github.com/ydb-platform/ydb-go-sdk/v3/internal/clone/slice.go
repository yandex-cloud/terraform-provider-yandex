package clone

func Int64Slice(src []int64) []int64 {
	if src == nil {
		return nil
	}

	dst := make([]int64, len(src))
	copy(dst, src)

	return dst
}
