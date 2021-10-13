package cairo_avl

import (
	"fmt"
	"math/big"
)

func balancedHeight(h_L, h_R *Felt) (h *Felt) {
	// Precondition: abs(h_L - h_R) <= 1
	if h_L.Cmp(h_R) == 0 {
		return new(Felt).Add(h_L, big.NewInt(1))
	} else if h_L.Cmp(new(Felt).Add(h_R, big.NewInt(1))) == 0 {
		return new(Felt).Add(h_L, big.NewInt(1))
	} else if h_R.Cmp(new(Felt).Add(h_L, big.NewInt(1))) == 0 {
		return new(Felt).Add(h_R, big.NewInt(1))
	} else {
		// balancedHeight is used with unbalanced trees
		//panic("balancedHeight: L and R trees non balanced")
		return NewFelt(0).Add(MaxBigInt(h_L, h_R), NewFelt(1))
	}
}

func rotateLeft(k, v *Felt, T_L, T_R, T_N *Node) (h *Felt, T *Node) {
	k_R, v_R, T_RL, T_RR, T_RN := exposeNode(T_R)
	h_L := height(T_L)
	h_RL := height(T_RL)
	h_RR := height(T_RR)
	h_dash := balancedHeight(h_L, h_RL)
	T_dash := makeNode(k, v, h_dash, T_L, T_RL, T_N)
	h = balancedHeight(h_dash, h_RR)
	return h, makeNode(k_R, v_R, h, T_dash, T_RR, T_RN)
}

func rotateRight(k, v *Felt, T_L, T_R, T_N *Node) (h *Felt, T *Node) {
	k_L, v_L, T_LL, T_LR, T_LN := exposeNode(T_L)
	h_R := height(T_R)
	h_LL := height(T_LL)
	h_LR := height(T_LR)
	h_dash := balancedHeight(h_LR, h_R)
	T_dash := makeNode(k, v, h_dash, T_LR, T_R, T_N)
	h = balancedHeight(h_dash, h_LL)
	return h, makeNode(k_L, v_L, h, T_LL, T_dash, T_LN)
}

func joinLeft(k, v *Felt, T_L, T_R, T_N *Node) (h *Felt, T *Node) {
	k_R, v_R, T_RL, T_RR, T_RN := exposeNode(T_R)
	h_RL := height(T_RL)
	h_L := height(T_L)
	h_RR := height(T_RR)
	if h_RL.Cmp(new(Felt).Add(h_L, big.NewInt(1))) <= 0 {
		h_dash := balancedHeight(h_L, h_RL)
		if h_dash.Cmp(new(Felt).Add(h_RR, big.NewInt(1))) <= 0 {
			h = balancedHeight(h_RR, h_dash)
			return h, makeNode(k_R, v_R, h, makeNode(k, v, h_dash, T_L, T_RL, T_N), T_RR, T_RN)
		} else {
			_, T_dash := rotateLeft(k, v, T_L, T_RL, T_N)
			return rotateRight(k_R, v_R, T_dash, T_RR, T_RN)
		}
	} else {
		h_dash, T_dash := joinLeft(k, v, T_L, T_RL, T_N)
		if h_dash.Cmp(new(Felt).Add(h_RR, big.NewInt(1))) <= 0 {
			h := balancedHeight(h_dash, h_RR)
			return h, makeNode(k_R, v_R, h, T_dash, T_RR, T_RN)
		} else {
			return rotateLeft(k_R, v_R, T_dash, T_RR, T_RN)
		}
	}
}

func joinRight(k, v *Felt, T_L, T_R, T_N *Node) (h *Felt, T *Node) {
	k_L, v_L, T_LL, T_LR, T_LN := exposeNode(T_L)
	h_LR := height(T_LR)
	h_R := height(T_R)
	h_LL := height(T_LL)
	if h_LR.Cmp(new(Felt).Add(h_R, big.NewInt(1))) <= 0 {
		h_dash := balancedHeight(h_LR, h_R)
		if h_dash.Cmp(new(Felt).Add(h_LL, big.NewInt(1))) <= 0 {
			h = balancedHeight(h_LL, h_dash)
			return h, makeNode(k_L, v_L, h, T_LL, makeNode(k, v, h_dash, T_LR, T_R, T_N), T_LN)
		} else {
			_, T_dash := rotateRight(k, v, T_LR, T_R, T_N)
			return rotateLeft(k_L, v_L, T_LL, T_dash, T_LN)
		}
	} else {
		h_dash, T_dash := joinRight(k, v, T_LR, T_R, T_N)
		if h_dash.Cmp(new(Felt).Add(h_LL, big.NewInt(1))) <= 0 {
			h := balancedHeight(h_dash, h_LL)
			return h, makeNode(k_L, v_L, h, T_LL, T_dash, T_LN)
		} else {
			return rotateLeft(k_L, v_L, T_LL, T_dash, T_LN)
		}
	}
}

func join(k, v *Felt, D_U, D_D *Dict, T_L, T_R, T_N *Node) (T *Node) {
	h_L := height(T_L)
	h_R := height(T_R)
	fmt.Println("join: h_L=", h_L, " h_R=", h_R)
	fmt.Println("join: T_N=", T_N.WalkKeysInOrder(), "D_D=", dictToNode(D_D).WalkKeysInOrder(), "D_U=", dictToNode(D_U).WalkKeysInOrder())
	N := Union(Difference(T_N, D_D), D_U)
	fmt.Println("join: N=", N.WalkKeysInOrder())
	if h_L.Cmp(new(Felt).Add(h_R, big.NewInt(1))) > 0 {
		_, T = joinRight(k, v, T_L, T_R, N)
		fmt.Println("join: T=", T)
		return T
	} else if h_R.Cmp(new(Felt).Add(h_L, big.NewInt(1))) > 0 {
		_, T = joinLeft(k, v, T_L, T_R, N)
		fmt.Println("join: T=", T)
		return T
	} else {
		h := balancedHeight(h_L, h_R)
		fmt.Println("join: h=", h)
		T = makeNode(k, v, h, T_L, T_R, N)
		fmt.Printf("join: T p=%p %+v k=%d\n", T, T, k.Uint64())
		return T
	}
}

func join2(T_L, T_R *Node) (T *Node) {
	if T_L == nil {
		return T_R
	} else {
		T_L_dash, k, v, N := splitLast(T_L)
		return join(k, v, nil, nil, T_L_dash, T_R, N)
	}
}

func split(T *Node, k *Felt) (T_L, T_R, T_N *Node) {
	if T == nil {
		return nil, nil, nil
	} else {
		m, v, L, R, N := exposeNode(T)
		if k.Cmp(m) == 0 {
			return L, R, N
		} else if k.Cmp(m) < 0 {
			L_L, L_R, L_N := split(L, k)
			return L_L, join(m, v, nil, nil, L_R, R, N), L_N
		} else {
			R_L, R_R, R_N := split(R, k)
			return join(m, v, nil, nil, L, R_L, N), R_R, R_N
		}
	}
}

func splitLast(T0 *Node) (T1 *Node, k, v *Felt, N *Node) {
	m, v, L, R, N := exposeNode(T0)
	if R == nil {
		return L, m, v, N
	} else {
		T_dash, k_dash, v_dash, N_dash := splitLast(R)
		return join(m, v, nil, nil, L, T_dash, N), k_dash, v_dash, N_dash
	}
}

func Insert(T *Node, k, v *Felt, N *Node) *Node {
	T_L, T_R, T_N := split(T, k)
	U_N := Union(N, nodeToDict(T_N))
	return join(k, v, nil, nil, T_L, T_R, U_N)
}

func Union(T0 *Node, D *Dict) (T1 *Node) {
	if T0 == nil {
		return dictToNode(D)
	} else if D == nil {
		return T0
	} else {
		k, v, D_L, D_R, D_U, D_D := exposeDict(D)
		fmt.Println("Union k=", k, "D_U=", dictToNode(D_U).WalkKeysInOrder(), " D_D=", dictToNode(D_D).WalkKeysInOrder())
		T_L, T_R, T_N := split(T0, k)
		fmt.Println("Union T_L=", T_L.WalkKeysInOrder(), " T_R=", T_R.WalkKeysInOrder())
		fmt.Println("Union D_L=", dictToNode(D_L).WalkKeysInOrder(), " D_R=", dictToNode(D_R).WalkKeysInOrder())
		L := Union(T_L, D_L)
		fmt.Println("Union L=", L.WalkKeysInOrder())
		R := Union(T_R, D_R)
		fmt.Println("Union R=", R.WalkKeysInOrder(), "D_U=", dictToNode(D_U).WalkKeysInOrder(), " D_D=", dictToNode(D_D).WalkKeysInOrder())
		return join(k, v, D_U, D_D, L, R, T_N)
	}
}

func Difference(T0 *Node, D *Dict) (T1 *Node) {
	if T0 == nil {
		return nil
	} else if D == nil {
		return T0
	} else {
		k, _, D_L, D_R, _, _ := exposeDict(D)
		T_L, T_R, _ := split(T0, k)
		L := Difference(T_L, D_L)
		R := Difference(T_R, D_R)
		return join2(L, R)
	}
}
