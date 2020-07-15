package project

type ByteView struct {
	b []byte
}

//缓存值长度
func (v *ByteView) Len() int {
	return len(v.b)
}

func (v *ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	temp := make([]byte, len(b))

	copy(temp, b)

	return temp
}
