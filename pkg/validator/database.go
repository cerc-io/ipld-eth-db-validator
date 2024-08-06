// VulcanizeDB
// Copyright © 2022 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package validator

import (
	"errors"

	"github.com/ethereum/go-ethereum/ethdb"
)

var errNotSupported = errors.New("this operation is not supported")

type database struct {
	ethdb.Database
}

func newDatabase(db ethdb.Database) *database {
	return &database{
		Database: db,
	}
}

func (d *database) Put(key []byte, value []byte) error {
	return errNotSupported
}

func (d *database) Delete(key []byte) error {
	return errNotSupported
}
