// Statping
// Copyright (C) 2018.  Hunter Long and the project contributors
// Written by Hunter Long <info@socialeck.com> and the project contributors
//
// https://github.com/hunterlong/statping
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

// Service is the main struct for Services
type Service struct {
	Id                  int64              `gorm:"primary_key;column:id" json:"id"`
	Name                string             `gorm:"column:name" json:"name"`
	Domain              string             `gorm:"column:domain" json:"domain"`
	Expected            NullString         `gorm:"column:expected" json:"expected"`
	ExpectedStatus      int                `gorm:"default:200;column:expected_status" json:"expected_status"`
	Interval            int                `gorm:"default:30;column:check_interval" json:"check_interval"`
	Type                string             `gorm:"column:check_type" json:"type"`
	Method              string             `gorm:"column:method" json:"method"`
	PostData            NullString         `gorm:"column:post_data" json:"post_data"`
	Port                int                `gorm:"not null;column:port" json:"port"`
	Timeout             int                `gorm:"default:30;column:timeout" json:"timeout"`
	Order               int                `gorm:"default:0;column:order_id" json:"order_id"`
	AllowNotifications  NullBool           `gorm:"default:true;column:allow_notifications" json:"allow_notifications"`
	VerifySSL           NullBool           `gorm:"default:false;column:verify_ssl" json:"verify_ssl"`
	Public              NullBool           `gorm:"default:true;column:public" json:"public"`
	GroupId             int                `gorm:"default:0;column:group_id" json:"group_id"`
	Headers             NullString         `gorm:"column:headers" json:"headers"`
	Permalink           NullString         `gorm:"column:permalink" json:"permalink"`
	FailureThreshold    int                `gorm:"default:0;column:failure_threshold" json:"failure_threshold"`
	CreatedAt           time.Time          `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time          `gorm:"column:updated_at" json:"updated_at"`
	Online              bool               `gorm:"-" json:"online"`
	Latency             float64            `gorm:"-" json:"latency"`
	PingTime            float64            `gorm:"-" json:"ping_time"`
	Online24Hours       float32            `gorm:"-" json:"online_24_hours"`
	AvgResponse         string             `gorm:"-" json:"avg_response"`
	Running             chan bool          `gorm:"-" json:"-"`
	Checkpoint          time.Time          `gorm:"-" json:"-"`
	SleepDuration       time.Duration      `gorm:"-" json:"-"`
	LastResponse        string             `gorm:"-" json:"-"`
	UserNotified        bool               `gorm:"-" json:"-"` // True if the User was already notified about a Downtime
	UpdateNotify        bool               `gorm:"-" json:"-"` // This Variable is a simple copy of `core.CoreApp.UpdateNotify.Bool`
	DownText            string             `gorm:"-" json:"-"` // Contains the current generated Downtime Text
	SuccessNotified     bool               `gorm:"-" json:"-"` // Is 'true' if the user has already be informed that the Services now again available
	LastStatusCode      int                `gorm:"-" json:"status_code"`
	LastOnline          time.Time          `gorm:"-" json:"last_success"`
	Failures            []FailureInterface `gorm:"-" json:"failures,omitempty"`
	Checkins            []CheckinInterface `gorm:"-" json:"checkins,omitempty"`
	CurrentFailureCount int                `gorm:"-" json:"-"`
}

// BeforeCreate for Service will set CreatedAt to UTC
func (s *Service) BeforeCreate() (err error) {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
		s.UpdatedAt = time.Now().UTC()
	}
	return
}

type ServiceInterface interface {
	Select() *Service
	CheckQueue(bool)
	Check(bool)
	Create(bool) (int64, error)
	Update(bool) error
	Delete() error
}

// Start will create a channel for the service checking go routine
func (s *Service) Start() {
	s.Running = make(chan bool)
}

// Close will stop the go routine that is checking if service is online or not
func (s *Service) Close() {
	if s.IsRunning() {
		close(s.Running)
	}
}

// IsRunning returns true if the service go routine is running
func (s *Service) IsRunning() bool {
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
