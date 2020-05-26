package gommr

import (
	"math"
	// "errors"
	"math/big"
	"strings"
	"testing"

	// "encoding/hex"

	"fmt"
)

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
func (r *proofRes) String() string {
	return fmt.Sprintf("{index:%v,hash:%s, td:%v}", r.index, r.h.Hex(), r.td)
}
func (p *ProofElem) String() string {
	if p.Cat == 2 {
		return fmt.Sprintf("[Child,%s]", p.Res.String())
	} else if p.Cat == 1 {
		return fmt.Sprintf("[Node,%s,Right:%v]", p.Res.String(), p.Right)
	} else {
		return fmt.Sprintf("[Root,%s,LeafNum:%v]", p.Res.String(), p.LeafNum)
	}
}
func (p *ProofInfo) String() string {
	elems := make([]string, len(p.Elems))
	for i, v := range p.Elems {
		elems[i] = v.String()
	}
	return fmt.Sprintf("RootHash:%s \n,RootDiff:%v,LeafNum:%v \n,Elems:%s", p.RootHash.Hex(),
		p.RootDifficulty, p.LeafNumber, strings.Join(elems, "\n "))
}
func Test03(t *testing.T) {
	a, b := uint64(1), uint64(2)
	fmt.Println("1:", RlpHash(a), "2:", RlpHash(b))
	val := uint64(0x4029000000000000)
	fmt.Println("val:", val, "fval:", ByteToFloat64(Uint64ToBytes(val)))

	aa := [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	fmt.Println("aa", Hash_to_f64(BytesToHash(aa[:])))
	fmt.Println("finish")
}

func Test04(t *testing.T) {
	right_difficulty, root_difficulty := big.NewInt(int64(1000)), big.NewInt(int64(10000))
	lambda, C, leaf_number := uint64(50), float64(float64(50)/100.0), uint64(10)

	aa := [32]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	root_hash := BytesToHash(aa[:])
	fmt.Println("root_hash:", root_hash)

	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(root_difficulty, right_difficulty)).Float64()
	fmt.Println("r1:", r1, "r2:", r2)

	required_queries := uint64(vd_calculate_m(float64(lambda), C, r1, r2, leaf_number) + 1.0)

	fmt.Println("required_queries:", required_queries)

	weights := []float64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, uint64(i)})
		random := Hash_to_f64(h)
		r3, _ := new(big.Float).SetInt(root_difficulty).Float64()
		aggr_weight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, aggr_weight)
		fmt.Println("i:", i, "aggr_weight:", aggr_weight)
	}
	res := make(map[int]int, 10)
	for _, weight := range weights {
		index := int(weight * 10)
		res[index]++
	}
	fmt.Println(res)
	fmt.Println("finish")
}

func Test05(t *testing.T) {
	mmr := NewMMR()
	for i := 0; i < 10000; i++ {
		mmr.push(&Node{
			value:      BytesToHash(IntToBytes(i)),
			difficulty: big.NewInt(1000),
		})
	}
	right_difficulty := big.NewInt(1000)
	fmt.Println("leaf_number:", mmr.getLeafNumber(), "root_difficulty:", mmr.getRootDifficulty())
	proof, blocks, eblocks := mmr.CreateNewProof(right_difficulty)
	fmt.Println("blocks_len:", len(blocks), "blocks:", blocks, "eblocks:", len(eblocks))
	fmt.Println("proof:", proof)
	pBlocks, err := VerifyRequiredBlocks(blocks, proof.RootHash, proof.RootDifficulty, right_difficulty, proof.LeafNumber)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	b := proof.VerifyProof(pBlocks)
	fmt.Println("b:", b)
	fmt.Println("finish")
}
