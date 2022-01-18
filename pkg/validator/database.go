// Copyright 2020 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

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

// ModifyAncients is not supported.
func (d *database) ModifyAncients(fn func(ethdb.AncientWriteOp) error) (int64, error) {
	return 0, nil
}

func (d *database) TruncateAncients(n uint64) error {
	return nil
}

func (d *database) Sync() error {
	return d.ethDB.Sync()
}

func (d *database) NewBatch() ethdb.Batch {
	return d.ethDB.NewBatch()
}

func (d *database) ReadAncients(fn func(ethdb.AncientReader) error) (err error) {
	return d.ethDB.ReadAncients(fn)
}

func (d *database) Close() error {
	return d.ethDB.Close()
}
