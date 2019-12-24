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

package notifier

import (
	"fmt"

	"github.com/hunterlong/statping/types"
	"github.com/hunterlong/statping/utils"
)

// OnSave will trigger a notifier when it has been saved - Notifier interface
func OnSave(method string) {
	for _, comm := range AllCommunications {
		if isType(comm, new(Notifier)) {
			notifier := comm.(Notifier)
			if notifier.Select().Method == method {
				notifier.OnSave()
			}
		}
	}
}

// OnFailure will be triggered when a service is failing - BasicEvents interface
func OnFailure(s *types.Service, f *types.Failure) {
	if !s.AllowNotifications.Bool {
		return
	}

	// check if User wants to receive every Status Change
	if s.UpdateNotify {
		// send only if User hasn't been already notified about the Downtime
		if s.UserNotified {
			return
		}
		s.UserNotified = true
	}

	for _, comm := range AllCommunications {
		if isType(comm, new(BasicEvents)) && isEnabled(comm) && (s.Online || inLimits(comm)) {
			notifier := comm.(Notifier).Select()
			utils.Log(1, fmt.Sprintf("Sending [OnFailure] '%v' notification for service %v", notifier.Method, s.Name))
			comm.(BasicEvents).OnFailure(s, f)
		}
	}
}

// OnSuccess will be triggered when a service is successful - BasicEvents interface
func OnSuccess(s *types.Service) {
	if !s.AllowNotifications.Bool {
		return
	}

	// check if User wants to receive every Status Change
	if s.UpdateNotify {
		// send only if User hasn't been already notified about the Uptime
		if !s.UserNotified {
			return
		}
		s.UserNotified = false
	}

	for _, comm := range AllCommunications {
		if isType(comm, new(BasicEvents)) && isEnabled(comm) && (!s.Online || inLimits(comm)) {
			notifier := comm.(Notifier).Select()
			utils.Log(1, fmt.Sprintf("Sending [OnSuccess] '%v' notification for service %v", notifier.Method, s.Name))
			comm.(BasicEvents).OnSuccess(s)
		}
	}
}

// OnNewService is triggered when a new service is created - ServiceEvents interface
func OnNewService(s *types.Service) {
	for _, comm := range AllCommunications {
		if isType(comm, new(ServiceEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending new service notification for service %v", s.Name))
			comm.(ServiceEvents).OnNewService(s)
		}
	}
}

// OnUpdatedService is triggered when a service is updated - ServiceEvents interface
func OnUpdatedService(s *types.Service) {
	if !s.AllowNotifications.Bool {
		return
	}
	for _, comm := range AllCommunications {
		if isType(comm, new(ServiceEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending updated service notification for service %v", s.Name))
			comm.(ServiceEvents).OnUpdatedService(s)
		}
	}
}

// OnDeletedService is triggered when a service is deleted - ServiceEvents interface
func OnDeletedService(s *types.Service) {
	if !s.AllowNotifications.Bool {
		return
	}
	for _, comm := range AllCommunications {
		if isType(comm, new(ServiceEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending deleted service notification for service %v", s.Name))
			comm.(ServiceEvents).OnDeletedService(s)
		}
	}
}

// OnNewUser is triggered when a new user is created - UserEvents interface
func OnNewUser(u *types.User) {
	for _, comm := range AllCommunications {
		if isType(comm, new(UserEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending new user notification for user %v", u.Username))
			comm.(UserEvents).OnNewUser(u)
		}
	}
}

// OnUpdatedUser is triggered when a new user is updated - UserEvents interface
func OnUpdatedUser(u *types.User) {
	for _, comm := range AllCommunications {
		if isType(comm, new(UserEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending updated user notification for user %v", u.Username))
			comm.(UserEvents).OnUpdatedUser(u)
		}
	}
}

// OnDeletedUser is triggered when a new user is deleted - UserEvents interface
func OnDeletedUser(u *types.User) {
	for _, comm := range AllCommunications {
		if isType(comm, new(UserEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending deleted user notification for user %v", u.Username))
			comm.(UserEvents).OnDeletedUser(u)
		}
	}
}

// OnUpdatedCore is triggered when the CoreApp settings are saved - CoreEvents interface
func OnUpdatedCore(c *types.Core) {
	for _, comm := range AllCommunications {
		if isType(comm, new(CoreEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending updated core notification"))
			comm.(CoreEvents).OnUpdatedCore(c)
		}
	}
}

// OnStart is triggered when the Statping service has started
func OnStart(c *types.Core) {
	for _, comm := range AllCommunications {
		if isType(comm, new(CoreEvents)) && isEnabled(comm) && inLimits(comm) {
			comm.(CoreEvents).OnUpdatedCore(c)
		}
	}
}

// OnNewNotifier is triggered when a new notifier is loaded
func OnNewNotifier(n *Notification) {
	for _, comm := range AllCommunications {
		if isType(comm, new(NotifierEvents)) && isEnabled(comm) && inLimits(comm) {
			comm.(NotifierEvents).OnNewNotifier(n)
		}
	}
}

// OnUpdatedNotifier is triggered when a notifier has been updated
func OnUpdatedNotifier(n *Notification) {
	for _, comm := range AllCommunications {
		if isType(comm, new(NotifierEvents)) && isEnabled(comm) && inLimits(comm) {
			utils.Log(1, fmt.Sprintf("Sending updated notifier for %v", n.Id))
			comm.(NotifierEvents).OnUpdatedNotifier(n)
		}
	}
}
