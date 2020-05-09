package gommr

import (
	// "errors"
	"encoding/hex"
	"math/big"
	"sort"

	// "fmt"
	// "math/big"
	"bytes"

	"github.com/go-mmr/gommr/rlp"
	"golang.org/x/crypto/sha3"
)

const (
	c      = float64(0.5)
	lambda = uint64(50)
)

type Hash [32]byte

func (h *Hash) Hex() string { return hex.EncodeToString(h[:]) }
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-32:]
	}
	copy(h[32-len(b):], b)
}
func BytesToHash(b []byte) Hash {
	var a Hash
	a.SetBytes(b)
	return a
}
func RlpHash(x interface{}) (h Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
func equal_hash(h1, h2 Hash) bool {
	return bytes.Equal(h1[:], h2[:])
}

type Node struct {
	value      Hash
	difficulty *big.Int
	index      uint64 // position in array
}

func (n *Node) getHash() Hash {
	return n.value
}
func (n *Node) setHash(h Hash) {
	n.value = h
}
func (n *Node) getDifficulty() *big.Int {
	return new(big.Int).Set(n.difficulty)
}
func (n *Node) setDifficulty(td *big.Int) {
	n.difficulty = new(big.Int).Set(td)
}
func (n *Node) setIndex(i uint64) {
	n.index = i
}
func (n *Node) getIndex() uint64 {
	return n.index
}
func (n *Node) clone() *Node {
	return &Node{
		value:      n.value,
		difficulty: new(big.Int).Set(n.difficulty),
		index:      n.index,
	}
}

type MerkleProof struct {
	mmrSize uint64
	proofs  []Hash
}

func newMerkleProof(mmrSize uint64, proof []Hash) *MerkleProof {
	return &MerkleProof{
		mmrSize: mmrSize,
		proofs:  proof,
	}
}
func (m *MerkleProof) verify(root Hash, pos uint64, leaf_hash Hash) bool {
	peaks := get_peaks(m.mmrSize)
	height := 0
	for _, proof := range m.proofs {
		// verify bagging peaks
		if pos_in_peaks(pos, peaks) {
			if pos == peaks[len(peaks)-1] {
				leaf_hash = merge2(leaf_hash, proof)
			} else {
				leaf_hash = merge2(proof, leaf_hash)
				pos = peaks[len(peaks)-1]
			}
			continue
		}
		// verify merkle path
		pos_height, next_height := pos_height_in_tree(pos), pos_height_in_tree(pos+1)
		if next_height > pos_height {
			// we are in right child
			leaf_hash = merge2(proof, leaf_hash)
			pos += 1
		} else {
			leaf_hash = merge2(leaf_hash, proof)
			pos += parent_offset(height)
		}
		height += 1
	}
	return equal_hash(leaf_hash, root)
}

type mmr struct {
	values  []*Node
	curSize uint64
	leafNum uint64
}

func newMMR() *mmr {
	return &mmr{
		values:  make([]*Node, 0, 0),
		curSize: 0,
		leafNum: 0,
	}
}
func (m *mmr) getNode(pos uint64) *Node {
	if pos > m.curSize-1 {
		return nil
	}
	return m.values[pos]
}
func (m *mmr) getLeafNumber() uint64 {
	return m.leafNum
}
func (m *mmr) push(n *Node) *Node {
	height, pos := 0, m.curSize
	n.index = pos
	m.values = append(m.values, n)
	m.leafNum++
	for {
		if pos_height_in_tree(pos+1) > height {
			pos++
			// calculate pos of left child and right child
			left_pos := pos - parent_offset(height)
			right_pos := left_pos + sibling_offset(height)
			left, right := m.values[left_pos], m.values[right_pos]
			parent := merge(left, right)
			// for test
			if parent.getIndex() != pos {
				panic("index not match")
			}
			parent.setIndex(pos)
			m.values = append(m.values, parent)
			height++
		} else {
			break
		}
	}
	m.curSize = pos + 1
	return n
}
func (m *mmr) getRoot() Hash {
	if m.curSize == 0 {
		return Hash{0}
	}
	if m.curSize == 1 {
		return m.values[0].getHash()
	}
	rootNode := m.bagRHSPeaks2(0, get_peaks(m.curSize))
	if rootNode != nil {
		return rootNode.getHash()
	} else {
		return Hash{0}
	}
	// return m.bagRHSPeaks(0, get_peaks(m.curSize))
}
func (m *mmr) getRootDifficulty() *big.Int {
	if m.curSize == 0 {
		return nil
	}
	if m.curSize == 1 {
		return m.values[0].getDifficulty()
	}
	rootNode := m.bagRHSPeaks2(0, get_peaks(m.curSize))
	if rootNode != nil {
		return rootNode.getDifficulty()
	}
	return nil
}
func (m *mmr) bagRHSPeaks(pos uint64, peaks []uint64) Hash {
	rhs_peak_hashes := make([]Hash, 0, 0)
	for _, v := range peaks {
		if v > pos {
			rhs_peak_hashes = append(rhs_peak_hashes, m.values[v].getHash())
		}
	}
	for {
		if len(rhs_peak_hashes) <= 1 {
			break
		}
		last := len(rhs_peak_hashes) - 1
		right := rhs_peak_hashes[last]
		rhs_peak_hashes = rhs_peak_hashes[:last]
		last = len(rhs_peak_hashes) - 1
		left := rhs_peak_hashes[last]
		rhs_peak_hashes = rhs_peak_hashes[:last]
		rhs_peak_hashes = append(rhs_peak_hashes, merge2(right, left))
	}
	if len(rhs_peak_hashes) == 1 {
		return rhs_peak_hashes[0]
	} else {
		return Hash{0}
	}
}
func (m *mmr) bagRHSPeaks2(pos uint64, peaks []uint64) *Node {
	rhsPeakNodes := make([]*Node, 0, 0)
	for _, v := range peaks {
		if v > pos {
			rhsPeakNodes = append(rhsPeakNodes, m.values[v])
		}
	}
	for {
		if len(rhsPeakNodes) <= 1 {
			break
		}
		last := len(rhsPeakNodes) - 1
		right := rhsPeakNodes[last]
		rhsPeakNodes = rhsPeakNodes[:last]
		last = len(rhsPeakNodes) - 1
		left := rhsPeakNodes[last]
		rhsPeakNodes = rhsPeakNodes[:last]
		parent := merge(right, left)
		parent.setIndex(right.getIndex() + 1)
		rhsPeakNodes = append(rhsPeakNodes, parent)
	}
	if len(rhsPeakNodes) == 1 {
		return rhsPeakNodes[0]
	}
	return nil
}
func (m *mmr) genProof(pos uint64) *MerkleProof {
	proofs := make([]Hash, 0, 0)
	height := 0
	for {
		if pos < m.curSize {
			pos_height, next_height := pos_height_in_tree(pos), pos_height_in_tree(pos+1)
			if next_height > pos_height {
				// get left child sib
				sib_pos := pos - sibling_offset(height)
				// break if sib is out of mmr
				if sib_pos >= m.curSize {
					break
				}
				proofs = append(proofs, m.values[sib_pos].getHash())
				// goto parent node
				pos = pos + 1
			} else {
				// get right child
				sib_pos := pos + sibling_offset(height)
				// break if sib is out of mmr
				if sib_pos >= m.curSize {
					break
				}
				proofs = append(proofs, m.values[sib_pos].getHash())
				// goto parent node
				next_pos := pos + parent_offset(height)
				pos = next_pos
			}
			height += 1
		} else {
			break
		}
	}
	// now pos is peak of the mountain(because pos can't find a sibling)
	peak_pos := pos
	peaks := get_peaks(m.curSize)
	// bagging rhs peaks into one hash
	rhs_peak_hash := m.bagRHSPeaks(peak_pos, peaks)
	if !equal_hash(rhs_peak_hash, Hash{0}) {
		proofs = append(proofs, rhs_peak_hash)
	}
	// put left peaks to proof
	for i := len(peaks) - 1; i >= 0; i-- {
		p := peaks[i]
		if p < pos {
			proofs = append(proofs, m.values[p].getHash())
		}
	}
	return newMerkleProof(m.curSize, proofs)
}
func (m *mmr) genProof2(right_difficulty *big.Int) {
	root_hash := m.getRoot()
	r1, _ := new(big.Float).SetInt(right_difficulty).Float64()
	r2, _ := new(big.Float).SetInt(new(big.Int).Add(m.getRootDifficulty(), right_difficulty)).Float64()
	required_queries := uint64(vd_calculate_m(float64(lambda), c, r1, r2, m.getLeafNumber()) + 1.0)

	weights, blocks := []float64{}, []uint64{}
	for i := 0; i < int(required_queries); i++ {
		h := RlpHash([]interface{}{root_hash, i})
		random, _ := new(big.Float).SetInt(new(big.Int).SetBytes(h[:])).Float64()
		r3, _ := new(big.Float).SetInt(m.getRootDifficulty()).Float64()
		aggr_weight := cdf(random, vd_calculate_delta(r1, r3))
		weights = append(weights, aggr_weight)
	}
	sort.Float64s(weights)
	for _, v := range weights {
		b := m.getChildByAggrWeight(v)
		blocks = append(blocks, b)
	}
	// Pick up at specific sync point
	// Add extra blocks, which are used for syncing from an already available state
	// 1. block : first block of current 30_000 block interval
	// 2. block : first block of previous 30_000 block interval
	// 3. block : first block of third last 30_000 block interaval
	// 4. block : first block of fourth last 30_000 block interval
	// 5. block : first block of fiftf last 30_000 block interval
	// 6. block : first block of sixth last 30_000 block interval
	// 7. block : first block of seventh last 30_000 block interval
	// 8. block : first block of eighth last 30_000 block interval
	// 9. block : first block of ninth last 30_000 block interval
	// 10. block: first block of tenth last 30_000 block interval

	// let mut extra_blocks = vec![];
	// let mut current_block = ((mmr.get_leaf_number() - 1) / 30000) * 30000;
	// let mut added = 0;
	// while current_block > 30000 && added < 10 {
	//     blocks.push(current_block);
	//     extra_blocks.push(current_block);
	//     current_block -= 30000;
	//     added += 1;
	// }

	// let mut blocks_dup = blocks.clone();
	// blocks_dup.sort();
	// let proof = Proof::generate_proof(&mut mmr, &mut blocks);

	// (proof, blocks_dup, extra_blocks)
}
func (m *mmr) getChildByAggrWeightDisc(weight *big.Int) uint64 {
	aggr_weight, aggr_node_number, curr_tree_number := big.NewInt(0), uint64(0), m.leafNum
	for {
		if curr_tree_number > 1 {
			left_tree_number := curr_tree_number / 2
			if !IsPowerOfTwo(curr_tree_number) {
				left_tree_number = NextPowerOfTwo(curr_tree_number) / 2
			}
			n := m.getNode(GetNodeFromLeaf(aggr_node_number+left_tree_number) - 1)
			if n == nil {
				panic("wrong pos1")
			}
			left_tree_difficulty := n.getDifficulty()
			if weight.Cmp(new(big.Int).Add(aggr_weight, left_tree_difficulty)) >= 0 {
				// branch right
				aggr_node_number += left_tree_number
				left_root_node_number := GetNodeFromLeaf(aggr_node_number) - 1
				n1 := m.getNode(left_root_node_number)
				if n1 == nil {
					panic("wrong pos2")
				}
				aggr_weight = new(big.Int).Add(aggr_weight, n1.getDifficulty())
				curr_tree_number = curr_tree_number - left_tree_number
			} else {
				// branch left
				curr_tree_number = left_tree_number
			}
		} else {
			break
		}
	}
	return aggr_node_number
}
func (m *mmr) getChildByAggrWeight(weight float64) uint64 {
	root_weight := m.getRootDifficulty()
	v1, _ := new(big.Float).Mul(new(big.Float).SetInt(root_weight), big.NewFloat(weight)).Int64()
	weight_disc := big.NewInt(v1)
	return m.getChildByAggrWeightDisc(weight_disc)
}
