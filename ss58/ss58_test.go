package ss58

import (
	"fmt"
	"testing"
)

func TestFormatAddress(t *testing.T) {
	out := "st7RTkrBasi9EMMyX9LyavzBEHGjp2xNcLy5LZaHA5xYU1CEC"
	in := "0x334eb447396e30197d4d3e810ef93ac77564fd358f2d4edc20319b6ffbb33492"
	addrType := 2254

	addr := Encode(in, addrType)
	fmt.Println("addr:", addr)

	addr = Decode(out, addrType)
	fmt.Printf("addr: 0x%s\n", addr)
}
