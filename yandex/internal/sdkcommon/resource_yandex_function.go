package sdkcommon

func MapBoolModeToString(isReadOnly bool) string {
	if isReadOnly {
		return "ro"
	}
	return "rw"
}
