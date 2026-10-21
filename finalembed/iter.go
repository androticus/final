package finalembed

import "github.com/androticus/final/internal/core"

// Iterator is the pull-based iteration protocol used by Final foreach (GetNext / Current).
// Embedder host types implement Iterable; users need not import these names in their public API.
type Iterator = core.Iterator

// Iterable host types expose GetIter() for OpTryGetIter / foreach without registering
// a Final _tryGetIter method.
type Iterable = core.Iterable
