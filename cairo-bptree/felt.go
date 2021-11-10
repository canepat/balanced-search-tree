package cairo_bptree

import "fmt"

type Felt = uint64

func pointerValue(pointer *Felt) string {
	if pointer != nil {
		return fmt.Sprintf("%d", *pointer)
	} else {
		return "<nil>"
	}
}

func ptr2pte(pointers []*Felt) []Felt {
	pointees := make([]Felt, 0)
	for _, ptr := range pointers {
		if ptr != nil {
			pointees = append(pointees, *ptr)
		} else {
			break
		}
	}
	return pointees
}
