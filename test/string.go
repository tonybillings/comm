package test

import "math/rand"

func GetRandomString(length int) []byte {
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(65 + (rand.Uint32() % 26))
	}
	return data
}

func GetRandomCString(length int) []byte {
	data := GetRandomString(length + 1)
	data[length] = 0
	return data
}
