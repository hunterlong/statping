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
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/go-mail/mail"
	"github.com/hunterlong/statping/core/notifier"
	"github.com/hunterlong/statping/types"
	"github.com/hunterlong/statping/utils"
	"html/template"
	"time"
)

const (
	mainEmailTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xmlns="http://www.w3.org/1999/xhtml">
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>Statping email</title>
</head>
<body style="-webkit-text-size-adjust: none; box-sizing: border-box; color: #74787E; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; height: 100%; line-height: 1.4; margin: 0; width: 100% !important;" bgcolor="#F2F4F6">
    <style type="text/css">
        body {
            width: 100% !important;
            height: 100%;
            margin: 0;
            line-height: 1.4;
            background-color: #F2F4F6;
            color: #74787E;
            -webkit-text-size-adjust: none;
        }
        @media only screen and (max-width: 600px) {
            .email-body_inner {
                width: 100% !important;
            }
            .email-footer {
                width: 100% !important;
            }
        }
        
        @media only screen and (max-width: 500px) {
            .button {
                width: 100% !important;
            }
        }
    </style>
    <table class="email-wrapper" width="100%" cellpadding="0" cellspacing="0" style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; margin: 0; padding: 0; width: 100%;" bgcolor="#F2F4F6">
        <tr>
            <td align="center" style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; word-break: break-word;">
                <table class="email-content" width="100%" cellpadding="0" cellspacing="0" style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; margin: 0; padding: 0; width: 100%;">
                    <tr>
                        <td class="email-body" width="100%" cellpadding="0" cellspacing="0" style="-premailer-cellpadding: 0; -premailer-cellspacing: 0; border-bottom-color: #EDEFF2; border-bottom-style: solid; border-bottom-width: 1px; border-top-color: #EDEFF2; border-top-style: solid; border-top-width: 1px; box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; margin: 0; padding: 0; width: 100%; word-break: break-word;" bgcolor="#FFFFFF">
                            <table class="email-body_inner" align="center" width="570" cellpadding="0" cellspacing="0" style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; margin: 0 auto; padding: 0; width: 570px;" bgcolor="#FFFFFF">
                                <tr>
                                    <td class="content-cell" style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; padding: 35px; word-break: break-word;">
                                        <h1 style="box-sizing: border-box; color: #2F3133; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; font-size: 19px; font-weight: bold; margin-top: 0;" align="left">
{{ .Name }} is {{ if .Online }}Online{{else}}Offline{{end}}!
</h1>
                                        <p style="box-sizing: border-box; color: #74787E; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; font-size: 16px; line-height: 1.5em; margin-top: 0;" align="left">

{{ if .Online }}
Your Statping service <a target="_blank" href="{{.Domain}}">{{.Name}}</a> is back online. This service has been triggered with a HTTP status code of '{{.LastStatusCode}}' and is currently online based on your requirements. Your service was reported online at {{.CreatedAt}}. </p>
{{ else }}
Your Statping service <a target="_blank" href="{{.Domain}}">{{.Name}}</a> has been triggered with a HTTP status code of '{{.LastStatusCode}}' and is currently offline based on your requirements. This failure was created on {{.CreatedAt}}. </p>
{{ end }}

{{if .LastResponse }}
                                        <h1 style="box-sizing: border-box; color: #2F3133; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; font-size: 19px; font-weight: bold; margin-top: 0;" align="left">
Last Response</h1>
                                        <p style="box-sizing: border-box; color: #74787E; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; font-size: 16px; line-height: 1.5em; margin-top: 0;" align="left">
{{ .LastResponse }} </p> {{end}}
                                        <table class="body-sub" style="border-top-color: #EDEFF2; border-top-style: solid; border-top-width: 1px; box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; margin-top: 25px; padding-top: 25px;">
                                            <td style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; word-break: break-word;"> <a href="/service/{{.Id}}" class="button button--blue" target="_blank" style="-webkit-text-size-adjust: none; background: #3869D4; border-color: #3869d4; border-radius: 3px; border-style: solid; border-width: 10px 18px; box-shadow: 0 2px 3px rgba(0, 0, 0, 0.16); box-sizing: border-box; color: #FFF; display: inline-block; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; text-decoration: none;">View Service</a> </td>
                                            <td style="box-sizing: border-box; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; word-break: break-word;"> <a href="/dashboard" class="button button--blue" target="_blank" style="-webkit-text-size-adjust: none; background: #3869D4; border-color: #3869d4; border-radius: 3px; border-style: solid; border-width: 10px 18px; box-shadow: 0 2px 3px rgba(0, 0, 0, 0.16); box-sizing: border-box; color: #FFF; display: inline-block; font-family: Arial, 'Helvetica Neue', Helvetica, sans-serif; text-decoration: none;">Statping Dashboard</a> </td>
                                        </table>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`
)

var (
	mailer *mail.Dialer
)

type email struct {
	*notifier.Notification
}

var emailer = &email{&notifier.Notification{
	Method:      "email",
	Title:       "email",
	Description: "Send emails via SMTP when services are online or offline.",
	Author:      "Hunter Long",
	AuthorUrl:   "https://github.com/hunterlong",
	Icon:        "far fa-envelope",
	Form: []notifier.NotificationForm{{
		Type:        "text",
		Title:       "SMTP Host",
		Placeholder: "Insert your SMTP Host here.",
		DbField:     "Host",
	}, {
		Type:        "text",
		Title:       "SMTP Username",
		Placeholder: "Insert your SMTP Username here.",
		DbField:     "Username",
	}, {
		Type:        "password",
		Title:       "SMTP Password",
		Placeholder: "Insert your SMTP Password here.",
		DbField:     "Password",
	}, {
		Type:        "number",
		Title:       "SMTP Port",
		Placeholder: "Insert your SMTP Port here.",
		DbField:     "Port",
	}, {
		Type:        "text",
		Title:       "Outgoing Email Address",
		Placeholder: "outgoing@email.com",
		DbField:     "Var1",
	}, {
		Type:        "email",
		Title:       "Send Alerts To",
		Placeholder: "sendto@email.com",
		DbField:     "Var2",
	}, {
		Type:        "text",
		Title:       "Disable TLS/SSL",
		Placeholder: "",
		SmallText:   "To Disable TLS/SSL insert 'true'",
		DbField:     "api_key",
	}},
}}

func init() {
	err := notifier.AddNotifier(emailer)
	if err != nil {
		panic(err)
	}
}

// Send will send the SMTP email with your authentication It accepts type: *emailOutgoing
func (u *email) Send(msg interface{}) error {
	email := msg.(*emailOutgoing)
	err := u.dialSend(email)
	if err != nil {
		return err
	}
	return nil
}

type emailOutgoing struct {
	To       string
	Subject  string
	Template string
	From     string
	Data     interface{}
	Source   string
	Sent     bool
}

// OnFailure will trigger failing service
func (u *email) OnFailure(s *types.Service, f *types.Failure) {
	email := &emailOutgoing{
		To:       u.Var2,
		Subject:  fmt.Sprintf("Service %v is Failing", s.Name),
		Template: mainEmailTemplate,
		Data:     interface{}(s),
		From:     u.Var1,
	}
	u.AddQueue(fmt.Sprintf("service_%v", s.Id), email)
}

// OnSuccess will trigger successful service
func (u *email) OnSuccess(s *types.Service) {
	if !s.Online || !s.SuccessNotified {
		var msg string
		if s.UpdateNotify {
			s.UpdateNotify = false
		}
		msg = s.DownText

		u.ResetUniqueQueue(fmt.Sprintf("service_%v", s.Id))
		email := &emailOutgoing{
			To:       u.Var2,
			Subject:  msg,
			Template: mainEmailTemplate,
			Data:     interface{}(s),
			From:     u.Var1,
		}
		u.AddQueue(fmt.Sprintf("service_%v", s.Id), email)
	}
}

func (u *email) Select() *notifier.Notification {
	return u.Notification
}

// OnSave triggers when this notifier has been saved
func (u *email) OnSave() error {
	utils.Log(1, fmt.Sprintf("Notification %v is receiving updated information.", u.Method))
	// Do updating stuff here
	return nil
}

// OnTest triggers when this notifier has been saved
func (u *email) OnTest() error {
	testService := &types.Service{
		Id:             1,
		Name:           "Example Service",
		Domain:         "https://www.youtube.com/watch?v=-u6DvRyyKGU",
		ExpectedStatus: 200,
		Interval:       30,
		Type:           "http",
		Method:         "GET",
		Timeout:        20,
		LastStatusCode: 200,
		Expected:       types.NewNullString("test example"),
		LastResponse:   "<html>this is an example response</html>",
		CreatedAt:      time.Now().Add(-24 * time.Hour),
	}
	email := &emailOutgoing{
		To:       u.Var2,
		Subject:  fmt.Sprintf("Service %v is Back Online", testService.Name),
		Template: mainEmailTemplate,
		Data:     testService,
		From:     u.Var1,
	}
	return u.dialSend(email)
}

func (u *email) dialSend(email *emailOutgoing) error {
	mailer = mail.NewDialer(emailer.Host, emailer.Port, emailer.Username, emailer.Password)
	emailSource(email)
	m := mail.NewMessage()
	// if email setting TLS is Disabled
	if u.ApiKey == "true" {
		mailer.SSL = false
	} else {
		mailer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	m.SetHeader("From", email.From)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	m.SetBody("text/html", email.Source)
	if err := mailer.DialAndSend(m); err != nil {
		utils.Log(3, fmt.Sprintf("email '%v' sent to: %v (size: %v) %v", email.Subject, email.To, len([]byte(email.Source)), err))
		return err
	}
	return nil
}

func emailSource(email *emailOutgoing) {
	source := emailTemplate(email.Template, email.Data)
	email.Source = source
}

func emailTemplate(contents string, data interface{}) string {
	t := template.New("email")
	t, err := t.Parse(contents)
	if err != nil {
		utils.Log(3, err)
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		utils.Log(2, err)
	}
	result := tpl.String()
	return result
}
