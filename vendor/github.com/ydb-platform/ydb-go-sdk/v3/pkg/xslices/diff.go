package xslices

func Diff[T any](from, to []T, cmp func(lhs, rhs T) int) (steady, added, dropped []T) {
	steady = make([]T, 0, min(len(to), len(from)))
	added = make([]T, 0, len(to))
	dropped = make([]T, 0, len(from))

	to = SortCopy(to, cmp)
	from = SortCopy(from, cmp)

	for i, j := 0, 0; ; {
		switch {
		case i == len(from) && j == len(to):
			return steady, added, dropped
		case i == len(from) && j < len(to):
			added = append(added, to[j])
			j++
		case i < len(from) && j == len(to):
			dropped = append(dropped, from[i])
			i++
		default:
			cmp := cmp(from[i], to[j])
			switch {
			case cmp < 0:
				dropped = append(dropped, from[i])
				i++
			case cmp > 0:
				added = append(added, to[j])
				j++
			default:
				steady = append(steady, from[i])
				i++
				j++
			}
		}
	}
}
