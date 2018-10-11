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

package notifier

import (
	"github.com/hunterlong/statup/source"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	dir    string
	METHOD = "example"
)

var service = &types.Service{
	Name:           "Interpol - All The Rage Back Home",
	Domain:         "https://www.youtube.com/watch?v=-u6DvRyyKGU",
	ExpectedStatus: 200,
	Interval:       30,
	Type:           "http",
	Method:         "GET",
	Timeout:        20,
}

var failure = &types.Failure{
	Issue: "testing",
}

var user = &types.User{
	Username: "admin",
	Email:    "info@email.com",
}

var core = &types.Core{
	Name: "testing notifiers",
}

func init() {
	dir = utils.Directory
	source.Assets()
	utils.InitLogs()
	injectDatabase()
}

func injectDatabase() {
	utils.DeleteFile(dir + "/statup.db")
	db, _ = gorm.Open("sqlite3", dir+"/statup.db")
	db.CreateTable(&Notification{})
}

func TestIsBasicType(t *testing.T) {
	assert.True(t, isType(example, new(Notifier)))
	assert.True(t, isType(example, new(BasicEvents)))
	assert.True(t, isType(example, new(ServiceEvents)))
	assert.True(t, isType(example, new(UserEvents)))
	assert.True(t, isType(example, new(CoreEvents)))
	assert.True(t, isType(example, new(NotifierEvents)))
	assert.True(t, isType(example, new(Tester)))
}

func TestLoad(t *testing.T) {
	notifiers := Load()
	assert.Equal(t, 1, len(notifiers))
}

func TestIsInDatabase(t *testing.T) {
	in := isInDatabase(example.Notification)
	assert.True(t, in)
}

func TestSelectNotification(t *testing.T) {
	notifier, err := SelectNotification(example)
	assert.Nil(t, err)
	assert.Equal(t, "example", notifier.Method)
	assert.False(t, notifier.Enabled)
	assert.False(t, notifier.IsRunning())
}

func TestAddQueue(t *testing.T) {
	msg := "this is a test in the queue!"
	example.AddQueue(msg)
	assert.Equal(t, 1, len(example.Queue))
	example.AddQueue(msg)
	assert.Equal(t, 2, len(example.Queue))
	example.AddQueue(msg)
	assert.Equal(t, 3, len(example.Queue))
	example.AddQueue(msg)
	assert.Equal(t, 4, len(example.Queue))
	example.AddQueue(msg)
	assert.Equal(t, 5, len(example.Queue))
}

func TestNotification_Update(t *testing.T) {
	notifier, err := SelectNotification(example)
	assert.Nil(t, err)
	notifier.Host = "http://demo.statup.io/api"
	notifier.Port = 9090
	notifier.Username = "admin"
	notifier.Password = "password123"
	notifier.Var1 = "var1_is_here"
	notifier.Var2 = "var2_is_here"
	notifier.ApiKey = "USBdu82HDiiuw9327yGYDGw"
	notifier.ApiSecret = "PQopncow929hUIDHGwiud"
	notifier.Limits = 10
	_, err = Update(example, notifier)
	assert.Nil(t, err)

	selected, err := SelectNotification(example)
	assert.Nil(t, err)
	assert.Equal(t, "http://demo.statup.io/api", selected.GetValue("host"))
	assert.Equal(t, "http://demo.statup.io/api", example.Notification.Host)
	assert.Equal(t, "http://demo.statup.io/api", example.Host)
	assert.Equal(t, "USBdu82HDiiuw9327yGYDGw", selected.GetValue("api_key"))
	assert.Equal(t, "USBdu82HDiiuw9327yGYDGw", example.ApiKey)
	assert.False(t, selected.Enabled)
	assert.False(t, selected.IsRunning())
}

func TestEnableNotification(t *testing.T) {
	notifier, err := SelectNotification(example)
	assert.Nil(t, err)
	notifier.Enabled = true
	updated, err := Update(example, notifier)
	assert.Nil(t, err)
	assert.True(t, updated.Enabled)
	assert.True(t, updated.IsRunning())
}

func TestIsEnabled(t *testing.T) {
	assert.True(t, isEnabled(example))
}

func TestIsRunning(t *testing.T) {
	assert.True(t, example.IsRunning())
}

func TestLastSent(t *testing.T) {
	notifier, err := SelectNotification(example)
	assert.Nil(t, err)
	assert.Equal(t, "0s", notifier.LastSent().String())
}

func TestWithinLimits(t *testing.T) {
	notifier, err := SelectNotification(example)
	assert.Nil(t, err)
	assert.Equal(t, 10, notifier.Limits)
	assert.True(t, inLimits(example))
}

func TestNotification_GetValue(t *testing.T) {
	notifier, err := SelectNotification(example)
	assert.Nil(t, err)
	val := notifier.GetValue("Host")
	assert.Equal(t, "http://demo.statup.io/api", val)
}

func TestOnSave(t *testing.T) {
	err := example.OnSave()
	assert.Equal(t, "onsave triggered", err.Error())
}

func TestOnSuccess(t *testing.T) {
	OnSuccess(service)
	assert.Equal(t, 7, len(example.Queue))
}

func TestOnFailure(t *testing.T) {
	OnFailure(service, failure)
	assert.Equal(t, 8, len(example.Queue))
}

func TestOnNewService(t *testing.T) {
	OnNewService(service)
	assert.Equal(t, 9, len(example.Queue))
}

func TestOnUpdatedService(t *testing.T) {
	OnUpdatedService(service)
	assert.Equal(t, 10, len(example.Queue))
}

func TestOnDeletedService(t *testing.T) {
	OnDeletedService(service)
	assert.Equal(t, 11, len(example.Queue))
}

func TestOnNewUser(t *testing.T) {
	OnNewUser(user)
	assert.Equal(t, 12, len(example.Queue))
}

func TestOnUpdatedUser(t *testing.T) {
	OnUpdatedUser(user)
	assert.Equal(t, 13, len(example.Queue))
}

func TestOnDeletedUser(t *testing.T) {
	OnDeletedUser(user)
	assert.Equal(t, 14, len(example.Queue))
}

func TestOnUpdatedCore(t *testing.T) {
	OnUpdatedCore(core)
	assert.Equal(t, 15, len(example.Queue))
}

func TestOnUpdatedNotifier(t *testing.T) {
	OnUpdatedNotifier(example.Select())
	assert.Equal(t, 16, len(example.Queue))
}

func TestRunAllQueueAndStop(t *testing.T) {
	assert.True(t, example.IsRunning())
	assert.Equal(t, 16, len(example.Queue))
	go Queue(example)
	time.Sleep(13 * time.Second)
	assert.NotZero(t, len(example.Queue))
	example.close()
	assert.False(t, example.IsRunning())
	assert.NotZero(t, len(example.Queue))
}
