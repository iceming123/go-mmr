package gommr

import (
	// "errors"
	"encoding/hex"
	// "fmt"
	// "math/big"
	"bytes"

	"github.com/go-mmr/gommr/rlp"
	"golang.org/x/crypto/sha3"
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
	value Hash
	index uint64
}

func (n *Node) getHash() Hash {
	return n.value
}
func (n *Node) setHash(h Hash) {
	n.value = h
}
func (n *Node) clone() *Node {
	return &Node{
		value: n.value,
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
	values   []*Node
	cur_size uint64
}

//              14
//          /         \
//         6          13
//       /   \       /   \
//      2     5     9     12     17
//     / \   /  \  / \   /  \   /  \
//    0   1 3   4 7   8 10  11 15  16 18

//      2
//     / \
//    0   1 3

func new_mmr() *mmr {
	return &mmr{
		values:   make([]*Node, 0, 0),
		cur_size: 0,
	}
}

func (m *mmr) push(n *Node) *Node {
	height, pos := 0, m.cur_size
	n.index = pos
	m.values = append(m.values, n)
	for pos_height_in_tree(pos+1) > height {
		pos++
		// calculate pos of left child and right child
		left_pos := pos - parent_offset(height)
		right_pos := left_pos + sibling_offset(height)
		left, right := m.values[left_pos], m.values[right_pos]
		parent := &Node{index: pos}
		merge(parent, left, right)
		m.values = append(m.values, parent)
		height++
	}
	m.cur_size = pos + 1
	return n
}
func (m *mmr) getRoot() Hash {
	if m.cur_size == 0 {
		return Hash{0}
	}
	if m.cur_size == 1 {
		return m.values[0].getHash()
	}
	return m.bag_rhs_peaks(0, get_peaks(m.cur_size))
}
func (m *mmr) bag_rhs_peaks(pos uint64, peaks []uint64) Hash {
	rhs_peak_hashes := make([]Hash, 0, 0)
	for _, v := range peaks {
		if v > pos {
			rhs_peak_hashes = append(rhs_peak_hashes, m.values[v].getHash())
		}
	}
	for len(rhs_peak_hashes) > 1 {
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
func (m *mmr) gen_proof(pos uint64) *MerkleProof {
	proofs := make([]Hash, 0, 0)
	height := 0
	for pos < m.cur_size {
		pos_height, next_height := pos_height_in_tree(pos), pos_height_in_tree(pos+1)
		if next_height > pos_height {
			// get left child sib
			sib_pos := pos - sibling_offset(height)
			// break if sib is out of mmr
			if sib_pos >= m.cur_size {
				break
			}
			proofs = append(proofs, m.values[sib_pos].getHash())
			// goto parent node
			pos = pos + 1
		} else {
			// get right child
			sib_pos := pos + sibling_offset(height)
			// break if sib is out of mmr
			if sib_pos >= m.cur_size {
				break
			}
			proofs = append(proofs, m.values[sib_pos].getHash())
			// goto parent node
			next_pos := pos + parent_offset(height)
			pos = next_pos
		}
		height += 1
	}
	// now pos is peak of the mountain(because pos can't find a sibling)
	peak_pos := pos
	peaks := get_peaks(m.cur_size)
	// bagging rhs peaks into one hash
	rhs_peak_hash := m.bag_rhs_peaks(peak_pos, peaks)
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
	return newMerkleProof(m.cur_size, proofs)
}
