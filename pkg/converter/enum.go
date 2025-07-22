package converter

type EnumI interface {
	~int32
	String() string
}

func EnumSliceToStrSlice[T EnumI](res []T) []string {
	var strs []string
	for _, v := range res {
		strs = append(strs, v.String())
	}
	return strs
}

func StrSliceToEnumSlice[T EnumI](enumMap map[string]int32, strs []string) []T {
	var results []T
	for _, str := range strs {
		results = append(results, T(enumMap[str]))
	}
	return results
}
