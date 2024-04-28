package utils

// InSlice 检查是否在切片中
func InSlice[T comparable](v T, sl []T) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}
