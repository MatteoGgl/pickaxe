package vault

// Reader is a thin VaultReader implementation that delegates to the
// package-level ListFiles and ReadFile functions, re-reading from disk each call.
type Reader struct{ dir string }

// NewReader returns a Reader backed by dir.
func NewReader(dir string) *Reader { return &Reader{dir: dir} }

func (r *Reader) ListFiles() ([]ResolvedFile, error)   { return ListFiles(r.dir) }
func (r *Reader) ReadFile(name string) (string, error) { return ReadFile(r.dir, name) }
