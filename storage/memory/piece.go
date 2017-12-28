package memory

import (
	"errors"
	"io"
	"sync"
	"time"
	// "math"

	"github.com/RoaringBitmap/roaring"
	"github.com/anacrolix/torrent/storage"
)

// Piece stores meta information about buffer contents
type Piece struct {
	c *Cache

	mu        *sync.Mutex
	Index     int
	Key       key
	Length    int64
	Position  int
	Hash      string
	Active    bool
	Completed bool
	Size      int64
	Accessed  time.Time

	Chunks *roaring.Bitmap
}

// Completion ...
func (p *Piece) Completion() storage.Completion {
	p.mu.Lock()
	defer p.mu.Unlock()

	return storage.Completion{
		Complete: p.Active && p.Completed && p.Size == p.Length && p.Length != 0,
		Ok:       true,
	}
}

// MarkComplete ...
func (p *Piece) MarkComplete() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Debugf("Complete: %#v", p.Index)
	p.Completed = true

	if !p.Active || p.Size != p.Length || p.Length == 0 {
		panic("piece is not completed")
	}

	return nil
}

// MarkNotComplete ...
func (p *Piece) MarkNotComplete() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Debugf("NotComplete: %#v", p.Index)
	p.Completed = false
	return nil
}

// GetBuffer ...
func (p *Piece) GetBuffer(iswrite bool) bool {
	p.c.mu.Lock()
	defer p.c.mu.Unlock()

	if p.Active {
		return true
	} else if p.Index >= len(p.c.pieces) {
		return false
	}

	if !p.Active && iswrite {
		for index, v := range p.c.positions {
			if v.Used {
				continue
			}

			v.Used = true
			v.Index = p.Index
			v.Key = p.Key

			p.Position = index
			p.Active = true
			p.Size = 0

			p.c.items[p.Key] = ItemState{}

			p.c.updateItem(p.Key, func(i *ItemState, ok bool) bool {
				if !ok {
					*i = p.GetState()
				}
				i.Accessed = time.Now()
				return ok
			})

			break
		}

		if !p.Active {
			log.Debugf("Buffer not assigned: %#v", p.c.positions)
			return false
		}
	}

	return true
}

// Seek File-like implementation
func (p *Piece) Seek(offset int64, whence int) (ret int64, err error) {
	log.Debugf("Seek lone: %#v", offset)
	return
}

// Write File-like implementation
func (p *Piece) Write(b []byte) (n int, err error) {
	log.Debugf("Write lone: %#v", len(b))
	return
}

// WriteAt File-like implementation
func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if buffered := p.GetBuffer(true); !buffered {
		log.Debugf("Can't get buffer write: %#v", p.Index)
		return 0, errors.New("Can't get buffer write")
	}

	chunkID, _ := p.GetChunkForOffset(off)
	p.Chunks.AddInt(chunkID)

	n = copy(p.c.buffers[p.Position][off:], b[:])

	p.Size += int64(n)
	p.onWrite()

	return
}

// Read File-like implementation
func (p *Piece) Read(b []byte) (n int, err error) {
	log.Debugf("Read lone: %#v", len(b))
	return
}

// ReadAt File-like implementation
func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if buffered := p.GetBuffer(false); !buffered {
		log.Debugf("No buffer read: %#v", p.Index)
		// return 0, nil
		return 0, io.EOF
	}

	requested := len(b)
	startIndex, _ := p.GetChunkForOffset(off)
	lastIndex, _ := p.GetChunkForOffset(off + int64(requested-chunkSize))

	if lastIndex < startIndex {
		lastIndex = startIndex
	}

	for i := startIndex; i <= lastIndex; i++ {
		if !p.Chunks.ContainsInt(i) {
			log.Debugf("Not contains read: %#v, Stats: %#v-%#v, Completed: %#v, Chunk: %#v (%#v-%#v), Request: %#v, len(%#v)", p.Index, p.Size, p.Length, p.Completed, i, startIndex, lastIndex, off, len(b))
			return 0, io.ErrUnexpectedEOF
		}
	}

	n = copy(b, p.c.buffers[p.Position][off:][:])
	if n != requested {
		log.Debugf("No matched read: %#v", p.Index)
		return 0, io.EOF
	}

	p.onRead()

	return n, nil
}

// GetChunkForOffset ...
func (p *Piece) GetChunkForOffset(offset int64) (index, margin int) {
	index = int(offset / chunkSize)
	margin = int(offset % chunkSize)

	return
}

// GetState ...
func (p *Piece) GetState() ItemState {
	return ItemState{
		Size:     p.Size,
		Accessed: p.Accessed,
	}
}

func (p *Piece) onRead() {
	p.c.mu.Lock()
	defer p.c.mu.Unlock()

	p.c.updateItem(p.Key, func(i *ItemState, ok bool) bool {
		i.Accessed = time.Now()
		return ok
	})
}

func (p *Piece) onWrite() {
	p.c.mu.Lock()
	defer p.c.mu.Unlock()

	p.c.updateItem(p.Key, func(i *ItemState, ok bool) bool {
		i.Accessed = time.Now()
		i.Size = p.Size
		return ok
	})
}

// Close File-like implementation
func (p *Piece) Close() error {
	return nil
}
