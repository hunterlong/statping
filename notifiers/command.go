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

package notifiers

import (
	"github.com/hunterlong/statping/core/notifier"
	"github.com/hunterlong/statping/types"
	"github.com/hunterlong/statping/utils"
	"time"
)

type commandLine struct {
	*notifier.Notification
}

var command = &commandLine{&notifier.Notification{
	Method:      "command",
	Title:       "Shell Command",
	Description: "Shell Command allows you to run customize shell or bash commands on the local machine it's running on.",
	Author:      "Hunter Long",
	AuthorUrl:   "https://github.com/hunterlong",
	Delay:       time.Duration(1 * time.Second),
	Icon:        "fas fa-terminal",
	Host:        "sh",
	Form: []notifier.NotificationForm{{
		Type:        "text",
		Title:       "Shell or Bash",
		Placeholder: "sh",
		DbField:     "host",
		SmallText:   "You can use 'sh', 'bash' or even an absolute path for an application",
	}, {
		Type:        "text",
		Title:       "Command to Run on OnSuccess",
		Placeholder: "ping google.com",
		DbField:     "var1",
		SmallText:   "This command will run everytime a service is receiving a Successful event.",
	}, {
		Type:        "text",
		Title:       "Command to Run on OnFailure",
		Placeholder: "ping offline.com",
		DbField:     "var2",
		SmallText:   "This command will run everytime a service is receiving a Failing event.",
	}}},
}

// init the command notifier
func init() {
	err := notifier.AddNotifier(command)
	if err != nil {
		panic(err)
	}
}

func runCommand(app, cmd string) (string, string, error) {
	outStr, errStr, err := utils.Command(cmd)
	return outStr, errStr, err
}

func (u *commandLine) Select() *notifier.Notification {
	return u.Notification
}

// OnFailure for commandLine will trigger failing service
func (u *commandLine) OnFailure(s *types.Service, f *types.Failure) {
	u.AddQueue(s.Id, u.Var2)
	u.Online = false
}

// OnSuccess for commandLine will trigger successful service
func (u *commandLine) OnSuccess(s *types.Service) {
	if !u.Online {
		u.ResetUniqueQueue(s.Id)
		u.AddQueue(s.Id, u.Var1)
	}
	u.Online = true
}

// OnSave for commandLine triggers when this notifier has been saved
func (u *commandLine) OnSave() error {
	u.AddQueue(0, u.Var1)
	u.AddQueue(0, u.Var2)
	return nil
}

// OnTest for commandLine triggers when this notifier has been saved
func (u *commandLine) OnTest() error {
	in, out, err := runCommand(u.Host, u.Var1)
	utils.Log(1, in)
	utils.Log(1, out)
	return err
}

// Send for commandLine will send message to expo command push notifications endpoint
func (u *commandLine) Send(msg interface{}) error {
	cmd := msg.(string)
	_, _, err := runCommand(u.Host, cmd)
	return err
}
