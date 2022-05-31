package validator

import (
	"github.com/ethereum/go-ethereum/ethdb"
)

type database struct {
	ethDB ethdb.Database
}

func newDatabase(db ethdb.Database) *database {
	return &database{
		ethDB: db,
	}
}

func (d *database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	return d.ethDB.NewIterator(prefix, start)
}

func (d *database) Has(key []byte) (bool, error) {
	return d.ethDB.Has(key)
}

func (d *database) Get(key []byte) ([]byte, error) {
	return d.ethDB.Get(key)
}

func (d *database) Put(key []byte, value []byte) error {
	return nil
}

func (d *database) Delete(key []byte) error {
	return nil
}

func (d *database) Stat(property string) (string, error) {
	return d.ethDB.Stat(property)
}

func (d *database) Compact(start []byte, limit []byte) error {
	return d.ethDB.Compact(start, limit)
}

// HasAncient returns an error as we don't have a backing chain freezer.
func (d *database) HasAncient(kind string, number uint64) (bool, error) {
	return d.ethDB.HasAncient(kind, number)
}

// Ancient returns an error as we don't have a backing chain freezer.
func (d *database) Ancient(kind string, number uint64) ([]byte, error) {
	return d.ethDB.Ancient(kind, number)
}

// AncientRange returns an error as we don't have a backing chain freezer.
func (d *database) AncientRange(kind string, start, max, maxByteSize uint64) ([][]byte, error) {
	return d.ethDB.AncientRange(kind, start, max, maxByteSize)
}

// Ancients returns an error as we don't have a backing chain freezer.
func (d *database) Ancients() (uint64, error) {
	return d.ethDB.Ancients()
}

// AncientSize returns an error as we don't have a backing chain freezer.
func (d *database) AncientSize(kind string) (uint64, error) {
	return d.ethDB.AncientSize(kind)
}

// Tail returns the number of first stored item in the freezer.
func (d *database) Tail() (uint64, error) {
	return d.Tail()
}

// ModifyAncients is not supported.
func (d *database) ModifyAncients(fn func(ethdb.AncientWriteOp) error) (int64, error) {
	return 0, nil
}

// TruncateHead discards all but the first n ancient data from the ancient store.
func (d *database) TruncateHead(n uint64) error {
	return nil
}

// TruncateTail discards the first n ancient data from the ancient store.
func (d *database) TruncateTail(n uint64) error {
	return nil
}

func (d *database) Sync() error {
	return d.ethDB.Sync()
}

// MigrateTable processes and migrates entries of a given table to a new format.
func (d *database) MigrateTable(string, func([]byte) ([]byte, error)) error {
	return nil
}

func (d *database) NewBatch() ethdb.Batch {
	return d.ethDB.NewBatch()
}

// NewBatchWithSize creates a write-only database batch with pre-allocated buffer.
func (d *database) NewBatchWithSize(size int) ethdb.Batch {
	return d.ethDB.NewBatchWithSize(size)
}

func (d *database) ReadAncients(fn func(ethdb.AncientReader) error) (err error) {
	return d.ethDB.ReadAncients(fn)
}

func (d *database) Close() error {
	return d.ethDB.Close()
}

// NewSnapshot creates a database snapshot based on the current state.
func (d *database) NewSnapshot() (ethdb.Snapshot, error) {
	return d.NewSnapshot()
}
