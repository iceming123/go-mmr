package gommr

import (
	"math/bits"
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
	bit_length := 64 - leadingZeros(pos);
    most_significant_bits := uint64(1) << uint64(bit_length - 1);
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
	return uint64(2) << uint64(height); 
}
func sibling_offset(height int) uint64 { 
	return (uint64(2) << uint64(height)) - 1; 
}
func merge(parent,left,right *Node) {
	hashes := make([]Hash,0,0)
	hashes = append(append(hashes,left.getHash()),right.getHash())
	parent.setHash(RlpHash(hashes))
}
func merge2(left,right Hash) Hash {
	hashes := make([]Hash,0,0)
	hashes = append(append(hashes,left),right)
	return RlpHash(hashes)
}
func left_peak_pos_by_height(height int) uint64 {
	return (uint64(1) << uint64(height + 1)) - 2
}
func left_peak_height_pos(mmrSize uint64) (int,uint64) {
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
    return height - 1,prev_pos
}
func get_right_peak(height int,peakPos,mmrSize uint64) (int,uint64) {
	//jump to right sibling
    peakPos += sibling_offset(height)
	//jump to left child
	for {
		if peakPos <= mmrSize - 1 {
			break
		}
		if height == 0 {
			//no right peak exists
			return height,0
		}
		height -= 1
		peakPos -= parent_offset(height)
	}
	return height,peakPos
}
func get_peaks(mmrSize uint64) []uint64 {
	res := make([]uint64,0,0)
	height, pos := left_peak_height_pos(mmrSize)
	res = append(res,pos)
	for {
		if height <= 0 {
			break
		}
		height,pos = get_right_peak(height,pos,mmrSize)
		if height == 0 && pos == 0{
			break
		} 
		res = append(res,pos)
	}
    return res
}
func pos_in_peaks(pos uint64, peaks []uint64) bool {
	for _,v := range peaks {
		if v == pos {
			return true
		}
	}
	return false
}