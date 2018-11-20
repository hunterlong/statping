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

// Hit struct is a 'successful' ping or web response entry for a service.
type Hit struct {
	Id        int64     `gorm:"primary_key;column:id"`
	Service   int64     `gorm:"column:service"`
	Latency   float64   `gorm:"column:latency"`
	PingTime  float64   `gorm:"column:ping_time"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// DbConfig struct is used for the database connection and creates the 'config.yml' file
type DbConfig struct {
	DbConn      string `yaml:"connection"`
	DbHost      string `yaml:"host"`
	DbUser      string `yaml:"user"`
	DbPass      string `yaml:"password"`
	DbData      string `yaml:"database"`
	DbPort      int64  `yaml:"port"`
	ApiKey      string `yaml:"api_key"`
	ApiSecret   string `yaml:"api_secret"`
	Project     string `yaml:"-"`
	Description string `yaml:"-"`
	Domain      string `yaml:"-"`
	Username    string `yaml:"-"`
	Password    string `yaml:"-"`
	Email       string `yaml:"-"`
	Error       error  `yaml:"-"`
	Location    string `yaml:"location"`
	LocalIP     string `yaml:"-"`
}
