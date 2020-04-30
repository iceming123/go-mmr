package gommr

import (
	// "errors"
	"testing"
	// "encoding/hex"
	"fmt"
	"encoding/binary"
	"bytes"
)

func IntToBytes(n int) []byte {
    data := int64(n)
    bytebuf := bytes.NewBuffer([]byte{})
    binary.Write(bytebuf, binary.BigEndian, data)
    return bytebuf.Bytes()
}

func run_mmr(count int,proof_pos uint64)  {
	mmr := new_mmr()
	positions := make([]*Node,0,0)
	
	for i:=0;i<count;i++ {
		positions = append(positions,mmr.push(&Node{
			value:	BytesToHash(IntToBytes(i)),
		}))
	}
	merkle_root := mmr.getRoot()
	// proof
    pos := positions[proof_pos].index
    // generate proof for proof_elem
    proof := mmr.gen_proof(pos)
    // verify proof
	result := proof.verify(merkle_root, pos,positions[proof_pos].getHash())
	fmt.Println("result:",result)
}
func Test01(t *testing.T)  {
	run_mmr(100,30)
	fmt.Println("finish")
}