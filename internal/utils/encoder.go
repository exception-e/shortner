package utils

import (
	"strconv"

	"github.com/spaolacci/murmur3"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GetHash(link string) uint64 {
	return uint64(murmur3.Sum32([]byte(link)))
}

func EncodeBase62(hash uint64) string {
	if hash == 0 {
		return "0"
	}
	var byteArr []byte

	for hash > 0 {
		c := base62Alphabet[hash%62]
		byteArr = append([]byte{c}, byteArr...)
		hash = hash / 62
	}
	return string(byteArr)
}

func DecodeBase62(link string) uint64 {
	var num uint64 = 0
	for _, ch := range link {
		var num1 uint64 = 0
		if ch >= '0' && ch <= '9' {
			num1 = uint64(ch - '0')
		}
		if ch >= 'A' && ch <= 'Z' {
			num1 = uint64(ch) - 'A' + 10
		}
		if ch >= 'a' && ch <= 'z' {
			num1 = uint64(ch) - 'a' + 36
		}
		num = num*62 + num1
	}
	return num
}

func AddSalt(link string, count int) string {
	newShortLink := EncodeBase62(GetHash(link + strconv.Itoa(count)))
	return newShortLink
}
