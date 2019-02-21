package main

type BitSet []uint64

func NewBitSet(bits uint) BitSet {
	n := (bits + 63) / 64
	return make(BitSet, n)
}

func (b BitSet) Bits() uint {
	return uint(len(b)) * 64
}

func (b BitSet) Set(i uint) (changed bool) {
	j, k := i/64, i%64
	x := b[j] | (1 << k)
	changed = b[j] != x
	b[j] = x
	return
}

func (b BitSet) Clear(i uint) (changed bool) {
	j, k := i/64, i%64
	x := b[j] &^ (1 << k)
	changed = b[j] != x
	b[j] = x
	return
}

func (b BitSet) Test(i uint) bool {
	j, k := i/64, i%64
	return (b[j] & (1 << k)) != 0
}

func (b BitSet) Union(c BitSet) (changed bool) {
	for i, e := range c {
		x := b[i] | e
		changed = changed || b[i] != x
		b[i] = x
	}
	return
}

func (b BitSet) Range(f func(uint)) {
	for i, w := range b {
		for j := uint(0); j < 64; j++ {
			if w&(1<<j) != 0 {
				f(uint(i)*64 + j)
			}
		}
	}
}
