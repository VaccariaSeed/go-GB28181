package go_GB28181

import (
	"math"
	"sync"
)

var seqNoLock sync.Mutex
var seqNo uint32

func flushSeqNo() uint32 {
	seqNoLock.Lock()
	defer seqNoLock.Unlock()
	seqNo++
	defer func() {
		if seqNo > math.MaxUint32 {
			seqNo = 0
		}
	}()
	return seqNo
}
