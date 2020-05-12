package gommr

import (
	"math"
	"math/big"
	"math/bits"
	"sort"
)

func countZore(num uint64) int {
	return bits.UintSize - bits.OnesCount64(num)
}
func leadingZeros(num uint64) int {
	return bits.LeadingZeros64(num)
}
func allOnes(num uint64) bool {
	return num != 0 && countZore(num) == leadingZeros(num)
}
func jumpLeft(pos uint64) uint64 {
	bit_length := 64 - leadingZeros(pos)
	most_significant_bits := uint64(1) << uint64(bit_length-1)
	return pos - (most_significant_bits - 1)
}
func pos_height_in_tree(pos uint64) int {
	pos += 1
	for {
		if !allOnes(pos) {
			pos = jumpLeft(pos)
		} else {
			break
		}
	}
	return 64 - leadingZeros(pos) - 1
}
func parent_offset(height int) uint64 {
	return uint64(2) << uint64(height)
}
func sibling_offset(height int) uint64 {
	return (uint64(2) << uint64(height)) - 1
}
func merge(left, right *Node) *Node {
	parent := &Node{}
	hashes := make([]Hash, 0, 0)
	hashes = append(append(hashes, left.getHash()), right.getHash())
	parent.setHash(RlpHash(hashes))
	parent.setDifficulty(new(big.Int).Add(left.getDifficulty(), right.getDifficulty()))
	parent.setIndex(right.getIndex() + 1)
	return parent
}
func merge2(left, right Hash) Hash {
	hashes := make([]Hash, 0, 0)
	hashes = append(append(hashes, left), right)
	return RlpHash(hashes)
}
func left_peak_pos_by_height(height int) uint64 {
	return (uint64(1) << uint64(height+1)) - 2
}
func left_peak_height_pos(mmrSize uint64) (int, uint64) {
	height := 0
	prev_pos := uint64(0)
	pos := left_peak_pos_by_height(height)
	//increase height and get most left pos of tree
	//once pos is out of mmr_size we consider previous pos is left peak
	for {
		if pos >= mmrSize {
			break
		}
		height += 1
		prev_pos = pos
		pos = left_peak_pos_by_height(height)
	}
	return height - 1, prev_pos
}
func get_right_peak(height int, peakPos, mmrSize uint64) (int, uint64) {
	//jump to right sibling
	peakPos += sibling_offset(height)
	//jump to left child
	for {
		if peakPos <= mmrSize-1 {
			break
		}
		if height == 0 {
			//no right peak exists
			return height, 0
		}
		height -= 1
		peakPos -= parent_offset(height)
	}
	return height, peakPos
}
func get_peaks(mmrSize uint64) []uint64 {
	res := make([]uint64, 0, 0)
	height, pos := left_peak_height_pos(mmrSize)
	res = append(res, pos)
	for {
		if height <= 0 {
			break
		}
		height, pos = get_right_peak(height, pos, mmrSize)
		if height == 0 && pos == 0 {
			break
		}
		res = append(res, pos)
	}
	return res
}
func pos_in_peaks(pos uint64, peaks []uint64) bool {
	for _, v := range peaks {
		if v == pos {
			return true
		}
	}
	return false
}
func IsPowerOfTwo(n uint64) bool {
	return n > 0 && ((n & (n - 1)) == 0)
}
func NextPowerOfTwo(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}
func GetNodeFromLeaf(ln uint64) uint64 {
	position, remaining := uint64(0), ln
	for {
		if remaining == 0 {
			break
		}
		leftTreeLeafNumber := remaining
		if !IsPowerOfTwo(remaining) {
			leftTreeLeafNumber = NextPowerOfTwo(remaining) / 2
		}
		position += leftTreeLeafNumber + leftTreeLeafNumber - 1
		remaining = remaining - leftTreeLeafNumber
	}
	return position
}

// calculate logarithm of x for base b:
//
// y = log_2(x)/log_2(b)
//
func log_b_of_x(b, x float64) float64 {
	return math.Log2(x) / math.Log2(b)
}

// calculate how many independent queries m are required to have the specified security of lambda
// and always check the last specified block difficulty manually in variable difficulty setting
func vd_calculate_m(lambda, c, block_difficulty, total_difficulty float64, n uint64) float64 {
	numerator := -lambda - math.Log2(c*float64(n))

	x := 1.0 - (1.0 / (log_b_of_x(c, block_difficulty/total_difficulty)))
	// x is not allowed to be negative
	if big.NewFloat(x).Sign() == -1 {
		x = 0.0
	}
	denumerator := math.Log2(x)
	return numerator / denumerator
}

// delta in variable difficulty setting is the sum of difficulty checked with probability 1 in the
// end
func vd_calculate_delta(block_difficulty, total_difficulty float64) float64 {
	return block_difficulty / total_difficulty
}

//
//             y(ln(delta))
// f(y) = 1 - e
//
// The cdf takes into account, that the last delta blocks are manually checked
func cdf(y, delta float64) float64 {
	return 1.0 - math.Exp(y*math.Log(delta))
}

//////////////////////////////////////////////////////////////////
func SortAndRemoveRepeat(slc []uint64) []uint64 {
	sort.Slice(slc, func(i, j int) bool {
		return slc[i] < slc[j]
	})

	result := []uint64{}
	tempMap := map[uint64]byte{}

	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}
func SortAndRemoveRepeat2(slc []*ProofBlock) []*ProofBlock {
	return nil
}
func reverseProofRes(s []*ProofBlock) []*ProofBlock {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
func binary_search(sli []uint64, val uint64) int {
	return 0
}
func splitAt(sli []uint64, pos int) ([]uint64, []uint64) {
	return nil, nil
}

// Get depth of the MMR with a specified leaf_number
func get_depth(leaf_number uint64) int {
	depth := 64 - leadingZeros(leaf_number) - 1
	if !IsPowerOfTwo(leaf_number) {
		depth += 1
	}
	return depth
}

// calc leaf number from complete node number
func node_to_leaf_number(node_number uint64) uint64 {
	return (node_number + 1) / 2
}

func leaf_to_node_number(leaf_number uint64) uint64 {
	return (2 * leaf_number) - 1
}

func get_left_leaf_number(leaf_number uint64) uint64 {
	if IsPowerOfTwo(leaf_number) {
		return leaf_number / 2
	} else {
		return NextPowerOfTwo(leaf_number) / 2
	}
}
