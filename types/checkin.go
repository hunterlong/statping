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

// Checkin struct will allow an application to send a recurring HTTP GET to confirm a service is online
type Checkin struct {
	Id          int64     `gorm:"primary_key;column:id"`
	ServiceId   int64     `gorm:"index;column:service"`
	Name        string    `gorm:"column:name"`
	Interval    int64     `gorm:"column:check_interval"`
	GracePeriod int64     `gorm:"column:grace_period"`
	ApiKey      string    `gorm:"column:api_key"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	Running     chan bool `gorm:"-" json:"-"`
}

// CheckinHit is a successful response from a Checkin
type CheckinHit struct {
	Id        int64     `gorm:"primary_key;column:id"`
	Checkin   int64     `gorm:"index;column:checkin"`
	From      string    `gorm:"column:from_location"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// Start will create a channel for the checkin checking go routine
func (s *Checkin) Start() {
	s.Running = make(chan bool)
}

// Close will stop the checkin routine
func (s *Checkin) Close() {
	if s.IsRunning() {
		close(s.Running)
	}
}

// IsRunning returns true if the checkin go routine is running
func (s *Checkin) IsRunning() bool {
	if s.Running == nil {
		return false
	}
	select {
	case <-s.Running:
		return false
	default:
		return true
	}
}
