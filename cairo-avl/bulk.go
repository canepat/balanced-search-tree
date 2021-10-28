package cairo_avl

import (
	"math/big"

	log "github.com/sirupsen/logrus"
)

type Counters struct {
	ExposedCount	uint64
	HeightCount	uint64
}

func computeHeight(h_L, h_R *Felt) (h *Felt) {
	if h_L.Cmp(h_R) == 0 {
		return new(Felt).Add(h_L, big.NewInt(1))
	} else if h_L.Cmp(new(Felt).Add(h_R, big.NewInt(1))) == 0 {
		return new(Felt).Add(h_L, big.NewInt(1))
	} else if h_R.Cmp(new(Felt).Add(h_L, big.NewInt(1))) == 0 {
		return new(Felt).Add(h_R, big.NewInt(1))
	} else {
		// L and R are unbalanced trees
		return NewFelt(0).Add(MaxBigInt(h_L, h_R), NewFelt(1))
	}
}

func rotateLeft(k, v *Felt, T_L, T_R, T_N *Node, c *Counters) (h *Felt, T *Node) {
	k_R, v_R, T_RL, T_RR, T_RN := exposeNode(T_R, c)
	h_L := height(T_L, c)
	h_RL := height(T_RL, c)
	h_RR := height(T_RR, c)
	h_dash := computeHeight(h_L, h_RL)
	T_dash := makeNode(k, v, h_dash, T_L, T_RL, T_N)
	h = computeHeight(h_dash, h_RR)
	return h, makeNode(k_R, v_R, h, T_dash, T_RR, T_RN)
}

func rotateRight(k, v *Felt, T_L, T_R, T_N *Node, c *Counters) (h *Felt, T *Node) {
	k_L, v_L, T_LL, T_LR, T_LN := exposeNode(T_L, c)
	h_R := height(T_R, c)
	h_LL := height(T_LL, c)
	h_LR := height(T_LR, c)
	h_dash := computeHeight(h_LR, h_R)
	T_dash := makeNode(k, v, h_dash, T_LR, T_R, T_N)
	h = computeHeight(h_dash, h_LL)
	return h, makeNode(k_L, v_L, h, T_LL, T_dash, T_LN)
}

func joinLeft(k, v *Felt, T_L, T_R, T_N *Node, c *Counters) (h *Felt, T *Node) {
	k_R, v_R, T_RL, T_RR, T_RN := exposeNode(T_R, c)
	h_RL := height(T_RL, c)
	h_L := height(T_L, c)
	h_RR := height(T_RR, c)
	if h_RL.Cmp(new(Felt).Add(h_L, big.NewInt(1))) <= 0 {
		h_dash := computeHeight(h_L, h_RL)
		if h_dash.Cmp(new(Felt).Add(h_RR, big.NewInt(1))) <= 0 {
			h = computeHeight(h_RR, h_dash)
			return h, makeNode(k_R, v_R, h, makeNode(k, v, h_dash, T_L, T_RL, T_N), T_RR, T_RN)
		} else {
			_, T_dash := rotateLeft(k, v, T_L, T_RL, T_N, c)
			return rotateRight(k_R, v_R, T_dash, T_RR, T_RN, c)
		}
	} else {
		h_dash, T_dash := joinLeft(k, v, T_L, T_RL, T_N, c)
		if h_dash.Cmp(new(Felt).Add(h_RR, big.NewInt(1))) <= 0 {
			h := computeHeight(h_dash, h_RR)
			return h, makeNode(k_R, v_R, h, T_dash, T_RR, T_RN)
		} else {
			return rotateLeft(k_R, v_R, T_dash, T_RR, T_RN, c)
		}
	}
}

func joinRight(k, v *Felt, T_L, T_R, T_N *Node, c *Counters) (h *Felt, T *Node) {
	k_L, v_L, T_LL, T_LR, T_LN := exposeNode(T_L, c)
	h_LR := height(T_LR, c)
	h_R := height(T_R, c)
	h_LL := height(T_LL, c)
	if h_LR.Cmp(new(Felt).Add(h_R, big.NewInt(1))) <= 0 {
		h_dash := computeHeight(h_LR, h_R)
		if h_dash.Cmp(new(Felt).Add(h_LL, big.NewInt(1))) <= 0 {
			h = computeHeight(h_LL, h_dash)
			return h, makeNode(k_L, v_L, h, T_LL, makeNode(k, v, h_dash, T_LR, T_R, T_N), T_LN)
		} else {
			_, T_dash := rotateRight(k, v, T_LR, T_R, T_N, c)
			return rotateLeft(k_L, v_L, T_LL, T_dash, T_LN, c)
		}
	} else {
		h_dash, T_dash := joinRight(k, v, T_LR, T_R, T_N, c)
		if h_dash.Cmp(new(Felt).Add(h_LL, big.NewInt(1))) <= 0 {
			h := computeHeight(h_dash, h_LL)
			return h, makeNode(k_L, v_L, h, T_LL, T_dash, T_LN)
		} else {
			return rotateLeft(k_L, v_L, T_LL, T_dash, T_LN, c)
		}
	}
}

func join(k, v *Felt, D_U, D_D *Dict, T_L, T_R, T_N *Node, c *Counters) (T *Node) {
	h_L := height(T_L, c)
	h_R := height(T_R, c)
	log.Traceln("join: h_L=", h_L, " h_R=", h_R)
	log.Traceln("join: T_N=", T_N.WalkKeysInOrder(), "D_D=", dictToNode(D_D).WalkKeysInOrder(), "D_U=", dictToNode(D_U).WalkKeysInOrder())
	N := Union(Difference(T_N, D_D, c), D_U, c)
	log.Traceln("join: N=", N.WalkKeysInOrder())
	if h_L.Cmp(new(Felt).Add(h_R, big.NewInt(1))) > 0 {
		_, T = joinRight(k, v, T_L, T_R, N, c)
		log.Traceln("join: T=", T)
		return T
	} else if h_R.Cmp(new(Felt).Add(h_L, big.NewInt(1))) > 0 {
		_, T = joinLeft(k, v, T_L, T_R, N, c)
		log.Traceln("join: T=", T)
		return T
	} else {
		h := computeHeight(h_L, h_R)
		log.Traceln("join: h=", h)
		T = makeNode(k, v, h, T_L, T_R, N)
		log.Tracef("join: T p=%p %+v k=%d\n", T, T, k.Uint64())
		return T
	}
}

func join2(T_L, T_R *Node, c *Counters) (T *Node) {
	if T_L == nil {
		return T_R
	} else {
		T_L_dash, k, v, N := splitLast(T_L, c)
		return join(k, v, nil, nil, T_L_dash, T_R, N, c)
	}
}

func split(T *Node, k *Felt, c *Counters) (T_L, T_R, T_N *Node) {
	if T == nil {
		return nil, nil, nil
	} else {
		m, v, L, R, N := exposeNode(T, c)
		if k.Cmp(m) == 0 {
			return L, R, N
		} else if k.Cmp(m) < 0 {
			L_L, L_R, L_N := split(L, k, c)
			return L_L, join(m, v, nil, nil, L_R, R, N, c), L_N
		} else {
			R_L, R_R, R_N := split(R, k, c)
			return join(m, v, nil, nil, L, R_L, N, c), R_R, R_N
		}
	}
}

func splitLast(T0 *Node, c *Counters) (T1 *Node, k, v *Felt, N *Node) {
	m, v, L, R, N := exposeNode(T0, c)
	if R == nil {
		return L, m, v, N
	} else {
		T_dash, k_dash, v_dash, N_dash := splitLast(R, c)
		return join(m, v, nil, nil, L, T_dash, N, c), k_dash, v_dash, N_dash
	}
}

func Insert(T *Node, k, v *Felt, N *Node) *Node {
	c := &Counters{}
	T_L, T_R, T_N := split(T, k, c)
	U_N := Union(N, nodeToDict(T_N), c)
	return join(k, v, nil, nil, T_L, T_R, U_N, c)
}

func Union(T0 *Node, D *Dict, c *Counters) (T1 *Node) {
	if T0 == nil {
		return dictToNode(D)
	} else if D == nil {
		return T0
	} else {
		k, v, D_L, D_R, D_U, D_D := exposeDict(D)
		log.Traceln("Union k=", k, "D_U=", dictToNode(D_U).WalkKeysInOrder(), " D_D=", dictToNode(D_D).WalkKeysInOrder())
		T_L, T_R, T_N := split(T0, k, c)
		log.Traceln("Union T_L=", T_L.WalkKeysInOrder(), " T_R=", T_R.WalkKeysInOrder())
		log.Traceln("Union D_L=", dictToNode(D_L).WalkKeysInOrder(), " D_R=", dictToNode(D_R).WalkKeysInOrder())
		L := Union(T_L, D_L, c)
		log.Traceln("Union L=", L.WalkKeysInOrder())
		R := Union(T_R, D_R, c)
		log.Traceln("Union R=", R.WalkKeysInOrder(), "D_U=", dictToNode(D_U).WalkKeysInOrder(), " D_D=", dictToNode(D_D).WalkKeysInOrder())
		return join(k, v, D_U, D_D, L, R, T_N, c)
	}
}

func Difference(T0 *Node, D *Dict, c *Counters) (T1 *Node) {
	if T0 == nil {
		return nil
	} else if D == nil {
		return T0
	} else {
		k, _, D_L, D_R, _, _ := exposeDict(D)
		T_L, T_R, _ := split(T0, k, c)
		L := Difference(T_L, D_L, c)
		R := Difference(T_R, D_R, c)
		return join2(L, R, c)
	}
}
