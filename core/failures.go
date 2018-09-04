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

package core

import (
	"fmt"
	"github.com/ararog/timeago"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"strings"
	"time"
)

type Failure struct {
	*types.Failure
}

func (s *Service) CreateFailure(f *types.Failure) (int64, error) {
	f.CreatedAt = time.Now()
	f.Service = s.Id
	s.Failures = append(s.Failures, f)
	col := DbSession.Collection("failures")
	uuid, err := col.Insert(f)
	if err != nil {
		utils.Log(3, err)
		return 0, err
	}
	if uuid == nil {
		return 0, err
	}
	return uuid.(int64), err
}

func (s *Service) AllFailures() []*types.Failure {
	var fails []*types.Failure
	col := DbSession.Collection("failures").Find("service", s.Id).OrderBy("-id")
	err := col.All(&fails)
	if err != nil {
		utils.Log(3, fmt.Sprintf("Issue getting failures for service %v, %v", s.Name, err))
		return nil
	}
	return fails
}

func (u *Service) DeleteFailures() {
	_, err := DbSession.Exec(`DELETE FROM failures WHERE service = ?`, u.Id)
	if err != nil {
		utils.Log(3, fmt.Sprintf("failed to delete all failures: %v", err))
	}
	u.Failures = nil
}

func (s *Service) LimitedFailures() []*Failure {
	var failArr []*Failure
	col := DbSession.Collection("failures").Find("service", s.Id).OrderBy("-id").Limit(10)
	col.All(&failArr)
	return failArr
}

func reverseFailures(input []*Failure) []*Failure {
	if len(input) == 0 {
		return input
	}
	return append(reverseFailures(input[1:]), input[0])
}

func (f *Failure) Ago() string {
	got, _ := timeago.TimeAgoWithTime(time.Now(), f.CreatedAt)
	return got
}

func (f *Failure) Delete() error {
	col := DbSession.Collection("failures").Find("id", f.Id)
	return col.Delete()
}

func CountFailures() uint64 {
	col := DbSession.Collection("failures").Find()
	amount, err := col.Count()
	if err != nil {
		utils.Log(2, err)
		return 0
	}
	return amount
}

func (s *Service) TotalFailures() (uint64, error) {
	col := DbSession.Collection("failures").Find("service", s.Id)
	amount, err := col.Count()
	return amount, err
}

func (s *Service) TotalFailures24Hours() (uint64, error) {
	col := DbSession.Collection("failures").Find("service = ? AND created_at > ? ", s.Id, time.Now().Add(time.Hour * -24).Format("2006-01-02 15:04:05"))
	amount, err := col.Count()
	return amount, err
}

func (f *Failure) ParseError() string {
	err := strings.Contains(f.Issue, "connection reset by peer")
	if err {
		return fmt.Sprintf("Connection Reset")
	}
	err = strings.Contains(f.Issue, "operation timed out")
	if err {
		return fmt.Sprintf("HTTP Request Timed Out")
	}
	err = strings.Contains(f.Issue, "x509: certificate is valid")
	if err {
		return fmt.Sprintf("SSL Certificate invalid")
	}
	err = strings.Contains(f.Issue, "Client.Timeout exceeded while awaiting headers")
	if err {
		return fmt.Sprintf("Connection Timed Out")
	}
	err = strings.Contains(f.Issue, "no such host")
	if err {
		return fmt.Sprintf("Domain is offline or not found")
	}
	err = strings.Contains(f.Issue, "HTTP Status Code")
	if err {
		return fmt.Sprintf("Incorrect HTTP Status Code")
	}
	err = strings.Contains(f.Issue, "connection refused")
	if err {
		return fmt.Sprintf("Connection Failed")
	}
	err = strings.Contains(f.Issue, "can't assign requested address")
	if err {
		return fmt.Sprintf("Unable to Request Address")
	}
	err = strings.Contains(f.Issue, "no route to host")
	if err {
		return fmt.Sprintf("Domain is offline or not found")
	}
	return f.Issue
}
