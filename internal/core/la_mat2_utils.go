package core

//Helpers for 2D matrices.

// Wraps the indices of a 2D matrix on the frame's data stack.
// This is used in convolution operations treating the matrix
// as a torus with wrap-around edges.
func wrapMat2dOnFrame(f *RunFrame, rows, cols int) {
	ds := f.RP.DataStack
	col := int(ds.Pop().(Int))
	row := int(ds.Pop().(Int))
	ds.Push(Int(wrapIndex(row, rows)))
	ds.Push(Int(wrapIndex(col, cols)))
}

func wrapIndex(i, n int) int {
	i %= n
	if i < 0 {
		i += n
	}
	return i
}

type RowCol struct {
	rows int
	cols int
}

func (rc *RowCol) Rows() int {
	return rc.rows
}

func (rc *RowCol) Cols() int {
	return rc.cols
}

func (rc *RowCol) GetRowIter() *Range {
	return &Range{beg: 0, end: rc.rows - 1}
}

func (rc *RowCol) GetColIter() *Range {
	return &Range{beg: 0, end: rc.cols - 1}
}
