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
	"github.com/hunterlong/statup/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	newServiceId int64
)

func TestSelectHTTPService(t *testing.T) {
	services, err := CoreApp.SelectAllServices()
	assert.Nil(t, err)
	assert.Equal(t, 15, len(services))
	assert.Equal(t, "Google", services[0].Name)
	assert.Equal(t, "http", services[0].Type)
}

func TestSelectAllServices(t *testing.T) {
	services := CoreApp.Services
	for _, s := range services {
		service := s.(*Service)
		service.Check(false)
		assert.True(t, service.IsRunning())
		t.Logf("ID: %v %v\n", service.Id, service.Name)
	}
	assert.Equal(t, 15, len(services))
}

func TestServiceDowntime(t *testing.T) {
	service := SelectService(15)
	downtime := service.Downtime()
	assert.True(t, downtime.Minutes() > 0)
}

func TestSelectTCPService(t *testing.T) {
	services := CoreApp.Services
	assert.Equal(t, 15, len(services))
	service := SelectService(5)
	assert.NotNil(t, service)
	assert.Equal(t, "Google DNS", service.Name)
	assert.Equal(t, "tcp", service.Type)
}

func TestUpdateService(t *testing.T) {
	service := SelectService(1)
	service2 := SelectService(2)
	assert.Equal(t, "Google", service.Name)
	assert.Equal(t, "Statup Github", service2.Name)
	assert.True(t, service.Online)
	assert.True(t, service2.Online)
	service.Name = "Updated Google"
	service.Interval = 5
	err := service.Update(true)
	assert.Nil(t, err)
	// check if updating pointer array shutdown any other service
	service2 = SelectService(2)
	assert.True(t, service2.Online)
}

func TestUpdateAllServices(t *testing.T) {
	services, err := CoreApp.SelectAllServices()
	assert.Nil(t, err)
	for k, srv := range services {
		srv.Name = "Changed " + srv.Name
		srv.Interval = k + 3
		err := srv.Update(true)
		assert.Nil(t, err)
	}
}

func TestServiceHTTPCheck(t *testing.T) {
	service := SelectService(1)
	service.Check(true)
	assert.Equal(t, "Changed Updated Google", service.Name)
	assert.True(t, service.Online)
}

func TestCheckHTTPService(t *testing.T) {
	service := SelectService(1)
	assert.Equal(t, "Changed Updated Google", service.Name)
	assert.True(t, service.Online)
	assert.Equal(t, 200, service.LastStatusCode)
	assert.NotZero(t, service.Latency)
	assert.NotZero(t, service.PingTime)
}

func TestServiceTCPCheck(t *testing.T) {
	service := SelectService(5)
	service.Check(true)
	assert.Equal(t, "Changed Google DNS", service.Name)
	assert.True(t, service.Online)
}

func TestCheckTCPService(t *testing.T) {
	service := SelectService(5)
	assert.Equal(t, "Changed Google DNS", service.Name)
	assert.True(t, service.Online)
	assert.NotZero(t, service.Latency)
	assert.NotZero(t, service.PingTime)
}

func TestServiceOnline24Hours(t *testing.T) {
	since := time.Now().Add(-24 * time.Hour).Add(-10 * time.Minute)
	service := SelectService(1)
	assert.Equal(t, float32(100), service.OnlineSince(since))
	service2 := SelectService(5)
	assert.Equal(t, float32(100), service2.OnlineSince(since))
	service3 := SelectService(14)
	assert.True(t, service3.OnlineSince(since) > float32(49))
}

func TestServiceSmallText(t *testing.T) {
	service := SelectService(5)
	text := service.SmallText()
	assert.Contains(t, text, "Online since")
}

func TestServiceAvgUptime(t *testing.T) {
	since := time.Now().Add(-24 * time.Hour).Add(-10 * time.Minute)
	service := SelectService(1)
	assert.NotEqual(t, "0.00", service.AvgUptime(since))
	service2 := SelectService(5)
	assert.Equal(t, "100", service2.AvgUptime(since))
	service3 := SelectService(13)
	assert.NotEqual(t, "0", service3.AvgUptime(since))
	service4 := SelectService(15)
	assert.NotEqual(t, "0", service4.AvgUptime(since))
}

func TestServiceHits(t *testing.T) {
	service := SelectService(5)
	hits, err := service.Hits()
	assert.Nil(t, err)
	assert.True(t, len(hits) > 1400)
}

func TestServiceLimitedHits(t *testing.T) {
	service := SelectService(5)
	hits, err := service.LimitedHits()
	assert.Nil(t, err)
	assert.Equal(t, int(1024), len(hits))
}

func TestServiceTotalHits(t *testing.T) {
	service := SelectService(5)
	hits, err := service.TotalHits()
	assert.Nil(t, err)
	assert.NotZero(t, hits)
}

func TestServiceSum(t *testing.T) {
	service := SelectService(5)
	sum, err := service.Sum()
	assert.Nil(t, err)
	assert.NotZero(t, sum)
}

func TestCountOnline(t *testing.T) {
	amount := CoreApp.CountOnline()
	assert.True(t, amount >= 2)
}

func TestCreateService(t *testing.T) {
	s := ReturnService(&types.Service{
		Name:           "That'll do 🐢",
		Domain:         "https://www.youtube.com/watch?v=rjQtzV9IZ0Q",
		ExpectedStatus: 200,
		Interval:       3,
		Type:           "http",
		Method:         "GET",
		Timeout:        20,
	})
	var err error
	newServiceId, err = s.Create(false)
	assert.Nil(t, err)
	assert.NotZero(t, newServiceId)
	newService := SelectService(newServiceId)
	assert.Equal(t, "That'll do 🐢", newService.Name)
}

func TestViewNewService(t *testing.T) {
	newService := SelectService(newServiceId)
	assert.Equal(t, "That'll do 🐢", newService.Name)
}

func TestCreateFailingHTTPService(t *testing.T) {
	s := ReturnService(&types.Service{
		Name:           "Bad URL",
		Domain:         "http://localhost/iamnothere",
		ExpectedStatus: 200,
		Interval:       2,
		Type:           "http",
		Method:         "GET",
		Timeout:        5,
	})
	var err error
	newServiceId, err = s.Create(false)
	assert.Nil(t, err)
	assert.NotZero(t, newServiceId)
	newService := SelectService(newServiceId)
	assert.Equal(t, "Bad URL", newService.Name)
	t.Log("new service ID: ", newServiceId)
}

func TestServiceFailedCheck(t *testing.T) {
	service := SelectService(17)
	assert.Equal(t, "Bad URL", service.Name)
	service.Check(false)
	assert.Equal(t, "Bad URL", service.Name)
	assert.False(t, service.Online)
}

func TestCreateFailingTCPService(t *testing.T) {
	s := ReturnService(&types.Service{
		Name:     "Bad TCP",
		Domain:   "localhost",
		Port:     5050,
		Interval: 30,
		Type:     "tcp",
		Timeout:  5,
	})
	var err error
	newServiceId, err = s.Create(false)
	assert.Nil(t, err)
	assert.NotZero(t, newServiceId)
	newService := SelectService(newServiceId)
	assert.Equal(t, "Bad TCP", newService.Name)
	t.Log("new failing tcp service ID: ", newServiceId)
}

func TestServiceFailedTCPCheck(t *testing.T) {
	service := SelectService(newServiceId)
	service.Check(false)
	assert.Equal(t, "Bad TCP", service.Name)
	assert.False(t, service.Online)
}

func TestCreateServiceFailure(t *testing.T) {
	fail := &types.Failure{
		Issue:  "This is not an issue, but it would container HTTP response errors.",
		Method: "http",
	}
	service := SelectService(8)
	id, err := service.CreateFailure(fail)
	assert.Nil(t, err)
	assert.NotZero(t, id)
}

func TestDeleteService(t *testing.T) {
	service := SelectService(newServiceId)

	count, err := CoreApp.SelectAllServices()
	assert.Nil(t, err)
	assert.Equal(t, 18, len(count))

	err = service.Delete()
	assert.Nil(t, err)

	services := CoreApp.Services
	assert.Equal(t, 17, len(services))
}

func TestServiceCloseRoutine(t *testing.T) {
	s := ReturnService(new(types.Service))
	s.Name = "example"
	s.Domain = "https://google.com"
	s.Type = "http"
	s.Method = "GET"
	s.ExpectedStatus = 200
	s.Interval = 1
	s.Start()
	assert.True(t, s.IsRunning())
	t.Log(s.Checkpoint)
	t.Log(s.SleepDuration)
	go s.CheckQueue(false)
	t.Log(s.Checkpoint)
	t.Log(s.SleepDuration)
	time.Sleep(5 * time.Second)
	t.Log(s.Checkpoint)
	t.Log(s.SleepDuration)
	assert.True(t, s.IsRunning())
	s.Close()
	assert.False(t, s.IsRunning())
	s.Close()
	assert.False(t, s.IsRunning())
}

func TestServiceCheckQueue(t *testing.T) {
	s := ReturnService(new(types.Service))
	s.Name = "example"
	s.Domain = "https://google.com"
	s.Type = "http"
	s.Method = "GET"
	s.ExpectedStatus = 200
	s.Interval = 1
	s.Start()
	assert.True(t, s.IsRunning())
	go s.CheckQueue(false)

	go func() {
		time.Sleep(5 * time.Second)
		t.Log(s.Checkpoint)
		time.Sleep(6 * time.Second)
	}()

	time.Sleep(5 * time.Second)
	assert.True(t, s.IsRunning())
	s.Close()
	assert.False(t, s.IsRunning())
	s.Close()
	assert.False(t, s.IsRunning())
}

func TestDNScheckService(t *testing.T) {
	s := ReturnService(new(types.Service))
	s.Name = "example"
	s.Domain = "http://localhost:9000"
	s.Type = "http"
	s.Method = "GET"
	s.ExpectedStatus = 200
	s.Interval = 1
	amount, err := s.dnsCheck()
	assert.Nil(t, err)
	assert.NotZero(t, amount)
}
