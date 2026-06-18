package network

import (
    "bytes"
    "sync"
)

var snapshotBufferPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}