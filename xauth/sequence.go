package auth

import (
	"fmt"
	"math"
	"strconv"
)

const (
	/**
	 * number后缀
	 */
	SuffixUser = "8"

	/**
	 * number位数
	 */
	WidthUser = 8
)

func EncodeUserNumberWithSuffix(userId int, suffix int) (int64, error) {

	// number 升级.
	userIdStr := strconv.FormatInt(int64(userId), 10)
	if len(userIdStr) >= 10 {
		return int64(userId), nil
	}

	userNumber := fmt.Sprintf("%s%d", encode(userId, WidthUser), suffix)
	return strconv.ParseInt(userNumber, 10, 64)
}

// EncodeUserNumber 补位
func EncodeUserNumber(userId int) (int64, error) {

	// number 升级.
	userIdStr := strconv.FormatInt(int64(userId), 10)
	if len(userIdStr) >= 10 {
		return int64(userId), nil
	}

	userNumber := fmt.Sprintf("%s%s", encode(userId, WidthUser), SuffixUser)
	return strconv.ParseInt(userNumber, 10, 64)
}

func encode(id int, width int) string {
	maximum := int(math.Pow10(width)) - 1
	superscript := int(math.Log(float64(maximum)) / math.Log(2))

	mapbit := getMapBit(width)
	ret := exchange(id, mapbit)

	ret += maximum - int(math.Pow(2, float64(superscript))) + 1

	format := "%0" + strconv.Itoa(width) + "s"
	return fmt.Sprintf(format, strconv.Itoa(ret))
}

func exchange(raw int, bitmap []uint) int {
	count := len(bitmap)

	ret := 0
	sign := 0x1 << uint(count)
	raw |= sign

	for i, v := range bitmap {
		value := (raw >> uint(i)) & 0x1
		ret |= value << v
	}

	return ret
}

func getMapBit(width int) []uint {
	mapbit := map[int][]uint{
		4:  {10, 2, 11, 3, 0, 1, 9, 7, 12, 6, 4, 8, 5},
		5:  {4, 3, 13, 15, 7, 8, 6, 2, 1, 10, 5, 12, 0, 11, 14, 9},
		6:  {2, 7, 10, 9, 16, 3, 6, 8, 0, 4, 1, 12, 11, 13, 18, 5, 15, 17, 14},
		7:  {18, 0, 2, 22, 8, 3, 1, 14, 17, 12, 4, 19, 11, 9, 13, 5, 6, 15, 10, 16, 20, 7, 21},
		8:  {11, 8, 4, 0, 16, 14, 22, 7, 3, 5, 13, 18, 24, 25, 23, 10, 1, 12, 6, 21, 17, 2, 15, 9, 19, 20},
		10: {32, 3, 1, 28, 21, 18, 30, 7, 12, 22, 20, 13, 16, 15, 6, 17, 9, 25, 11, 8, 4, 27, 14, 31, 5, 23, 24, 29, 0, 10, 19, 26, 2},
		11: {9, 13, 2, 29, 11, 32, 14, 33, 24, 8, 27, 4, 22, 20, 5, 0, 21, 25, 17, 28, 34, 6, 23, 26, 30, 3, 7, 19, 16, 15, 12, 31, 1, 35, 10, 18},
		12: {31, 4, 16, 33, 35, 29, 17, 37, 12, 28, 32, 22, 7, 10, 14, 26, 0, 9, 8, 3, 20, 2, 13, 5, 36, 27, 23, 15, 19, 34, 38, 11, 24, 25, 30, 21, 18, 6, 1},
	}

	if v, ok := mapbit[width]; ok {
		return v
	}

	return []uint{}
}

func DecodeUserNumber(userNumber int64) (int, error) {
	suffixLength := len(SuffixUser)
	userNumberStr := strconv.FormatInt(userNumber, 10)
	// number升级.
	if len(userNumberStr) >= 10 {
		return int(userNumber), nil
	}
	rawNumber := userNumberStr[0 : len(userNumberStr)-suffixLength]
	number, err := strconv.ParseInt(rawNumber, 10, 64)
	if err != nil {
		return 0, err
	}
	return decode(number), nil
}

func decode(userNumber int64) int {
	width := len(strconv.FormatInt(userNumber, 10))
	maxNum := int(math.Pow10(width)) - 1
	superscript := int(math.Log(float64(maxNum)) / math.Log(2))

	raw := userNumber - int64(maxNum) + int64(math.Pow(2, float64(superscript))) - 1
	mapbit := getMapBit(width)
	decodeMap := arrayFlip(mapbit)
	return exchangeWithMap(int(raw), decodeMap)

}

// 键值互换
func arrayFlip(mapBit []uint) map[uint]uint {
	res := map[uint]uint{}
	for i, v := range mapBit {
		res[v] = uint(i)
	}
	return res
}

func exchangeWithMap(raw int, bitmap map[uint]uint) int {
	count := len(bitmap)

	ret := 0
	sign := 0x1 << uint(count)
	raw |= sign

	for i, v := range bitmap {
		value := (raw >> uint(i)) & 0x1
		ret |= value << v
	}

	return ret
}
