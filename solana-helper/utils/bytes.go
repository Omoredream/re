package Utils

import (
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

// https://solana.com/developers/courses/program-optimization/program-architecture#sizes

// bool

func BoolToBytes(data bool) (b []byte) {
	if data {
		b = []byte{1}
	} else {
		b = []byte{0}
	}
	return
}

func BytesToBool(b []byte) (data bool) {
	if len(b) != 1 {
		panic("数据长度不符合预期")
	}
	data = b[0] == 1
	return
}

// u8

func Uint8ToBytesL(n uint8) (b []byte) {
	b = make([]byte, 1)
	b[0] = n
	return
}

func BytesLToUint8(b []byte) (n uint8) {
	if len(b) != 1 {
		panic("数据长度不符合预期")
	}
	n = b[0]
	return
}

// u16

func Uint16ToBytesL(n uint16) (b []byte) {
	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, n)
	return
}

func BytesLToUint16(b []byte) (n uint16) {
	if len(b) != 2 {
		panic("数据长度不符合预期")
	}
	n = binary.LittleEndian.Uint16(b)
	return
}

// u32

func Uint32ToBytesL(n uint32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return
}

func BytesLToUint32(b []byte) (n uint32) {
	if len(b) != 4 {
		panic("数据长度不符合预期")
	}
	n = binary.LittleEndian.Uint32(b)
	return
}

// u64

func Uint64ToBytesL(n uint64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return
}

func BytesLToUint64(b []byte) (n uint64) {
	if len(b) != 8 {
		panic("数据长度不符合预期")
	}
	n = binary.LittleEndian.Uint64(b)
	return
}

// i8

func Int8ToBytesL(n int8) (b []byte) {
	return Uint8ToBytesL(uint8(n))
}

// i16

func Int16ToBytesL(n int16) (b []byte) {
	return Uint16ToBytesL(uint16(n))
}

// i32

func Int32ToBytesL(n int32) (b []byte) {
	return Uint32ToBytesL(uint32(n))
}

// i64

func Int64ToBytesL(n int64) (b []byte) {
	return Uint64ToBytesL(uint64(n))
}

// [T;amount]

func ArrayTToBytes[T any](to func(data T) (b []byte), data []T) (b []byte) {
	for i := range data {
		b = Append(b, to(data[i]))
	}
	return
}

func ByteArrayToBytes(data []byte) (b []byte) {
	//b = ArrayTToBytes(Uint8ToBytesL, data)
	b = data[:]
	return
}

// PubKey

func PublicKeyToBytes(data solana.PublicKey) (b []byte) {
	return data.Bytes()
}

// Vec<T>

func VecTToBytes[T any](to func(data T) (b []byte), data ...T) (b []byte) {
	if len(data) > 0xffffffff {
		panic("数据过长")
	}
	b = Append(Uint32ToBytesL(uint32(len(data))), ArrayTToBytes(to, data))
	return
}

// Vec<T> 特殊类型, alt 使用

func Vec64TToBytes[T any](to func(data T) (b []byte), data ...T) (b []byte) {
	if len(data) > 0xffffffff {
		panic("数据过长")
	}
	b = Append(Uint64ToBytesL(uint64(len(data))), ArrayTToBytes(to, data))
	return
}

// String

func StringToBytes(data string) (b []byte) {
	b = BytesToBytes([]byte(data))
	return
}

func BytesToBytes(data []byte) (b []byte) {
	//b = VecTToBytes(Uint8ToBytesL, data...)
	if len(data) > 0xffffffff {
		panic("数据过长")
	}
	b = Append(Uint32ToBytesL(uint32(len(data))), ByteArrayToBytes(data))
	return
}

// Option<T>

func OptionTToBytes[T any](to func(data T) (b []byte), data ...T) (b []byte) {
	option := data != nil && len(data) > 0
	b = Append(BoolToBytes(option))
	if option {
		b = Append(b, to(data[0]))
	}
	return
}

// Enum

// EnumToBytes 需要注意同个 Enum 的不同 index 的附加数据长度应一致, 以最长数据定义为准
func EnumToBytes(index uint8, data []byte) (b []byte) {
	b = Append(Uint8ToBytesL(index), data)
	return
}
