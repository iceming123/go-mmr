package gommr

import (
	"math"
	// "errors"
	"math/big"
	"testing"

	// "encoding/hex"
	"bytes"
	"encoding/binary"
	"fmt"
)

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func run_mmr(count int, proof_pos uint64) {
	mmr := newMMR()
	positions := make([]*Node, 0, 0)

	for i := 0; i < count; i++ {
		positions = append(positions, mmr.push(&Node{
			value:      BytesToHash(IntToBytes(i)),
			difficulty: big.NewInt(0),
		}))
	}
	merkle_root := mmr.getRoot()
	// proof
	pos := positions[proof_pos].index
	// generate proof for proof_elem
	proof := mmr.genProof(pos)
	// verify proof
	result := proof.verify(merkle_root, pos, positions[proof_pos].getHash())
	fmt.Println("result:", result)
}
func Test01(t *testing.T) {
	run_mmr(10000, 50)
	fmt.Println("finish")
}

func Test02(t *testing.T) {
	num := uint64(0)
	a := NextPowerOfTwo(num)
	b := float64(100)
	fmt.Println("b:", math.Log(b), "pos_height:", get_depth(6))
	fmt.Println("aa", a, "isPow:", IsPowerOfTwo(num), "GetNodeFromLeaf:", GetNodeFromLeaf(6))
}
func modify_slice(v []int) []int {
	fmt.Println("len(v):", len(v))
	v = append(v, 100)
	fmt.Println("len(v):", len(v))
	return v
}

func Test03(t *testing.T) {
	val := uint64(0x4029000000000000)
	fmt.Println("val:", val, "fval:", ByteToFloat64(Uint64ToBytes(val)))

	aa := [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	fmt.Println("aa", Hash_to_f64(BytesToHash(aa[:])))
	fmt.Println("finish")
}
