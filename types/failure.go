// Statup
// Copyright (C) 2018.  Hunter Long and the project contributors
// Written by Hunter Long <info@socialeck.com> and the project contributors
//
// https://github.com/hunterlong/statup
//
// The licenses for most software and other practical works are designed
// to take away your freedom to share and change the works.  By contrast,
// the GNU General Public License is intended to guarantee your freedom to
// share and change all versions of a program--to make sure it remains free
// software for all its users.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"time"
)

// Failure is a failed attempt to check a service. Any a service does not meet the expected requirements,
// a new Failure will be inserted into database.
type Failure struct {
	Id               int64     `gorm:"primary_key;column:id" json:"id"`
	Issue            string    `gorm:"column:issue" json:"issue"`
	Method           string    `gorm:"column:method" json:"method,omitempty"`
	MethodId         int64     `gorm:"column:method_id" json:"method_id,omitempty"`
	Service          int64     `gorm:"index;column:service" json:"-"`
	PingTime         float64   `gorm:"column:ping_time"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	FailureInterface `gorm:"-" json:"-"`
}

type FailureInterface interface {
	Ago() string        // Ago returns a human readble timestamp
	ParseError() string // ParseError returns a human readable error for a service failure
}
