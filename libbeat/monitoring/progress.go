package monitoring

import (
	"expvar"
	"os"
	"sync"

	"github.com/elastic/beats/v7/filebeat/input/file"
)

type File struct {
	Source string `json:"source"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

func newFile(source string, size, offset int64) File {
	return File{
		Source: source,
		Size:   size,
		Offset: offset,
	}
}

type RegistryProgress struct {
	sync.Mutex
	fileSlice []File
	fun       expvar.Func
}

func NewRegistryProgress(r *Registry, name string, opts ...Option) *RegistryProgress {
	if r == nil {
		r = Default
	}

	v := &RegistryProgress{
		fileSlice: []File{},
	}
	v.fun = func() interface{} {
		return v.fileSlice
	}
	addVar(r, name, opts, v, v.fun)
	return v
}

func (rp *RegistryProgress) Visit(_ Mode, vs Visitor) {
	switch vs.(type) {
	case *KeyValueVisitor:
		vs.OnInterface(rp.fun)
	case *structSnapshotVisitor:
		vs.OnInterface(rp.fileSlice)
	case *flatSnapshotVisitor:
		vs.OnInterface(rp.fileSlice)
	}
}

func (rp *RegistryProgress) Add(states []file.State) {
	rp.Lock()
	defer rp.Unlock()

	fileSlice := make([]File, 0)
	for _, state := range states {
		fileInfo, _ := os.Stat(state.Source)
		fileSlice = append(fileSlice, newFile(state.Source, fileInfo.Size(), state.Offset))
	}
	rp.fileSlice = fileSlice
}
