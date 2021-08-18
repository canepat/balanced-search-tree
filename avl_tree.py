# -*- coding: utf-8 -*-
"""
Adelson-Velsky and Landis (AVL) self-balancing binary search tree implementation. References:
- https://zhjwpku.com/assets/pdf/AED2-10-avl-paper.pdf
- https://www.cs.cmu.edu/~blelloch/papers/BFS16.pdf
- https://github.com/cmuparlay/PAM
"""

from __future__ import annotations
from typing import Optional

class AvlNode:
    key: int
    height: int
    parent: Optional[AvlNode]
    left: Optional[AvlNode]
    right: Optional[AvlNode]
    value: int

    def __init__(self: AvlNode, key: int, left: Optional[AvlNode] = None, right: Optional[AvlNode] = None) -> None:
        self.key = key
        self.left = left
        self.right = right
        if self.left:
            self.left.parent = self
        if self.right:
            self.right.parent = self
        self.height = self.compute_height()
        self.parent = None
        self.value = 0

    def compute_height(self: AvlNode) -> int:
        if self.left:
            if self.right:
                h = max(self.left.compute_height(), self.right.compute_height()) + 1
            else:
                h = self.left.compute_height() + 1
        else:
            if self.right:
                h = self.right.compute_height() + 1
            else:
                h = 0
        return h

    def path_depth(self: AvlNode) -> int:
        depth = 0
        ancestor = self.parent
        while ancestor:
            depth += 1
            ancestor = ancestor.parent
        return depth

    def walk_inorder(self: AvlNode) -> list[AvlNode]:
        left_nodes = self.left.walk_inorder() if self.left else []
        right_nodes = self.right.walk_inorder() if self.right else []
        return [self] + left_nodes + right_nodes

    def __repr__(self: AvlNode) -> str:
        dump = ''
        for node in self.walk_inorder():
            depth = node.path_depth()
            indent = ' ' * depth
            dump += f'{indent}{depth}) k={node.key} h={node.height} p={node.parent.key if node.parent else ""}\n'
        return dump

def graph(t: AvlNode, filename: str) -> None:
    colors = ['#FDF3D0', '#DCE8FA', '#D9E7D6', '#F1CFCD', '#F5F5F5', '#E1D5E7', '#FFE6CC', 'white']
    with open(filename + '.dot', 'w') as f:
        f.write('strict digraph {\n')
        f.write('node [shape=record];\n')
        for n in t.walk_inorder():
            left = ''
            right = ''
            if n.left:
                left = '<L>L'
            if n.right:
                right = '<R>R'
            f.write(f'{n.key} [label="{left}|{{<C>{n.key}|{n.value}}}|{right}" style=filled fillcolor="{colors[2]}"];\n')
            if n.parent:
                direction = 'L' if n.parent.left == n else 'R'
                f.write(f'{n.parent.key}:{direction} -> {n.key}:C;\n')
        f.write('}\n')

def graph_and_picture(t: AvlNode, filename: str) -> None:
    graph(t, filename)
    import subprocess
    subprocess.call(['dot', '-Tpng', filename + '.dot', '-o', filename + '.png'])

def expose(n: AvlNode) -> tuple[int, int, int]:
    return (n.left, n.key, n.right) if n else (None, None, None)

def height(n: AvlNode) -> int:
    return n.height if n else 0

def update(n: AvlNode) -> AvlNode:
    n.height = max(n.left.height if n.left else 0, n.right.height if n.right else 0) + 1
    return n

def rotate_left(n: AvlNode) -> AvlNode:
    y = n.right
    lsub = y.left
    y.left = n
    n.right = lsub
    update(y.left)
    update(y)
    return y

def rotate_right(n: AvlNode) -> AvlNode:
    y = n.left
    rsub = y.right
    y.right = n
    n.left = rsub
    update(y.right)
    update(y)
    return y

def double_rotate_left(n: AvlNode) -> AvlNode:
    r = n.right
    root = r.left
    r.left = root.right
    update(r)
    root.right = r
    n.right = root
    return rotate_left(n)

def double_rotate_right(n: AvlNode) -> AvlNode:
    l = n.left
    root = l.right
    l.right = root.left
    update(l)
    root.left = l
    n.left = root
    return rotate_right(n)

def is_left_heavy(left: AvlNode, right: AvlNode) -> bool:
    return height(left) > height(right) + 1

def is_single_rotation(n: AvlNode, dir: bool) -> bool:
    return height(n.left) > height(n.right) if dir else height(n.left) <= height(n.right)

"""
def join_right(left_tree: AvlTree, k: int, right_tree: AvlTree) -> AvlTree:
    l, k_dash, c = expose(left_tree.root)
    if height(c) <= height(right_tree.root) + 1:
        t_dash = node(c, k, right_tree)
        if height(t_dash.root) <= height(l) + 1:
            return node(l, k_dash, t_dash)
        else:
            return rotate_left(node(l, k_dash, rotate_right(t_dash)))
    else:
        t_dash = join_right(c, k, right_tree)
        t_double_dash = node(l, k_dash, t_dash)
        if height(t_dash) <= height(l) + 1:
            return t_double_dash
        else:
            return rotate_left(t_double_dash)
"""

def join_balanced(left: AvlNode, k: int, right: AvlNode) -> AvlNode:
    return AvlNode(k, left, right)

def join_right(n1: AvlNode, k: int, n2: AvlNode) -> AvlNode:
    if not is_left_heavy(n1, n2):
        return join_balanced(n1, k, n2)
    n = n1
    n.right = join_right(n.right, k, n2)
    n.right.parent = n
    if is_left_heavy(n.right, n.left):
        if is_single_rotation(n.right, False):
            n = rotate_left(n)
        else:
            n = double_rotate_left(n)
    else:
        update(n)
    return n

def join_left(n1: AvlNode, k: int, n2: AvlNode) -> AvlNode:
    if not is_left_heavy(n2, n1):
        return join_balanced(n1, k, n2)
    n = n2
    n.left = join_left(n1, k, n.left)
    n.left.parent = n
    if is_left_heavy(n.left, n.right):
        if is_single_rotation(n.left, True):
            n = rotate_right(n)
        else:
            n = double_rotate_right(n)
    else:
        update(n)
    return n

def join(n1: AvlNode, k: int, n2: AvlNode) -> AvlNode:
    if is_left_heavy(n1, n2):
        return join_right(n1, k, n2)
    if is_left_heavy(n2, n1):
        return join_left(n1, k, n2)
    return join_balanced(n1, k, n2)

def split(T: AvlNode, k: int) -> tuple[AvlNode, bool, AvlNode]:
    if T == None:
        return (None, False, None)
    L, m, R = expose(T)
    if k == m:
        return (L, True, R)
    elif k < m:
        L_l, b, L_r = split(L, k)
        return (L_l, b, join(L_r, m , R))
    else:
        R_l, b, R_r = split(R, k)
        return (join(L, m, R_l), b, R_r)

def split_last(T: AvlNode) -> tuple[AvlNode, int]:
    L, k, R = expose(T)
    if R == None:
        return (L, k)
    else:
        T_dash, k_dash = split_last(R)
        return (join(L, k, T_dash), k_dash)

def join2(Tl: AvlNode, Tr: AvlNode) -> AvlNode:
    if Tl == None:
        return Tr
    else:
        Tl_dash, k = split_last(Tl)
        return join(Tl_dash, k, Tr)

def insert(T: AvlNode, k: int):
    Tl, _, Tr = split(T, k)
    return join(Tl, k, Tr)

def delete(T: AvlNode, k: int):
    Tl, _, Tr = split(T, k)
    return join2(Tl, Tr)

def union(T1: AvlNode, T2: AvlNode) -> AvlNode:
    if T1 == None:
        return T2
    elif T2 == None:
        return T1
    else:
        L2, k2, R2 = expose(T2)
        L1, _, R1 = split(T1, k2)
        Tl = union(L1, L2) # parallelizable
        Tr = union(R1, R2) # parallelizable
        return join(Tl, k2, Tr)

def intersect(T1: AvlNode, T2: AvlNode) -> AvlNode:
    if T1 == None:
        return None
    elif T2 == None:
        return None
    else:
        L2, k2, R2 = expose(T2)
        L1, b, R1 = split(T1, k2)
        Tl = intersect(L1, L2) # parallelizable
        Tr = intersect(R1, R2) # parallelizable
        if b:
            return join(Tl, k2, Tr)
        else:
            return join2(Tl, Tr)

def difference(T1: AvlNode, T2: AvlNode) -> AvlNode:
    if T1 == None:
        return None
    elif T2 == None:
        return T1
    else:
        L2, k2, R2 = expose(T2)
        L1, _, R1 = split(T1, k2)
        Tl = difference(L1, L2) # parallelizable
        Tr = difference(R1, R2) # parallelizable
        return join2(Tl, Tr)

t1 = AvlNode(0)
print(f't1:\n{t1}\n')

t2 = AvlNode(18,
        AvlNode(15), AvlNode(21)
    )
print(f't2:\n{t2}\n')
t3 = AvlNode(188,
        AvlNode(155,
            AvlNode(154), AvlNode(156)
        ),
        AvlNode(210,
            AvlNode(200,
                AvlNode(199), AvlNode(202)),
            AvlNode(300,
                AvlNode(201), AvlNode(1560))
        )
    )
print(f't3:\n{t3}\n')
j1 = join(t2, 50, t3)
print(f'j1:\n{j1}\n')
print(f'j1 keys: {[n.key for n in j1.walk_inorder()]}\n')
graph_and_picture(j1, 'j1')

t4 = AvlNode(19,
        AvlNode(11), AvlNode(157)
    )
print('t4:\n' + str(t4) + '\n')
u1 = union(j1, t4)
print(f'u1:\n{u1}\n')
print(f'u1 keys: {[n.key for n in u1.walk_inorder()]}\n')
graph_and_picture(u1, 'u1')

t5 = AvlNode(4,
        AvlNode(1), AvlNode(5)
    )
print('t5:\n' + str(t5) + '\n')
t6 = AvlNode(3,
        AvlNode(2), AvlNode(7)
    )
print('t6:\n' + str(t6) + '\n')
u2 = union(t5, t6)
print(f'u2:\n{u2}\n')
print(f'u2 keys: {[n.key for n in u2.walk_inorder()]}\n')
graph_and_picture(u2, 'u2')

d1 = difference(u2, t5)
print(f'd1:\n{d1}\n')
print(f'd1 keys: {[n.key for n in d1.walk_inorder()]}\n')
graph_and_picture(d1, 'd1')
d2 = difference(u2, t6)
print(f'd2:\n{d2}\n')
print(f'd2 keys: {[n.key for n in d2.walk_inorder()]}\n')
graph_and_picture(d2, 'd2')
