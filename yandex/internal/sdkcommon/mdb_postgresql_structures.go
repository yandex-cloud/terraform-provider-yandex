package sdkcommon

func CopyMapIntString(source map[int]string) map[int]string {
	res := make(map[int]string)
	for k, v := range source {
		res[k] = v
	}
	return res
}
