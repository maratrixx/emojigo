package util

import "unsafe"

// 字符串转字节切片
func S2b(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// 字节切片转换字符串
func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
