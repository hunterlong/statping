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

package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hunterlong/statup/core"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"net"
	"net/http"
	"time"
)

func renderServiceChartsHandler(w http.ResponseWriter, r *http.Request) {
	services := core.CoreApp.Services
	w.Header().Set("Content-Type", "text/javascript")
	w.Header().Set("Cache-Control", "max-age=60")

	end := time.Now().UTC()
	start := time.Now().Add((-24 * 7) * time.Hour).UTC()
	var srvs []*core.Service
	for _, s := range services {
		srvs = append(srvs, s.(*core.Service))
	}
	out := struct {
		Services []*core.Service
		Start    int64
		End      int64
	}{srvs, start.Unix(), end.Unix()}

	executeJSResponse(w, r, "charts.js", out)
}

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	executeResponse(w, r, "services.html", core.CoreApp.Services, nil)
}

type serviceOrder struct {
	Id    int64 `json:"service"`
	Order int   `json:"order"`
}

func reorderServiceHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var newOrder []*serviceOrder
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&newOrder)
	for _, s := range newOrder {
		service := core.SelectService(s.Id)
		service.Order = s.Order
		service.Update(false)
	}
	w.WriteHeader(http.StatusOK)
}

func servicesViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fields := parseGet(r)
	r.ParseForm()

	startField := utils.StringInt(fields.Get("start"))
	endField := utils.StringInt(fields.Get("end"))
	group := r.Form.Get("group")
	serv := core.SelectService(utils.StringInt(vars["id"]))
	if serv == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	end := time.Now().UTC()
	start := end.Add((-24 * 7) * time.Hour).UTC()

	if startField != 0 {
		start = time.Unix(startField, 0).UTC()
	}
	if endField != 0 {
		end = time.Unix(endField, 0).UTC()
	}
	if group == "" {
		group = "hour"
	}

	data := core.GraphDataRaw(serv, start, end, group, "latency")

	out := struct {
		Service   *core.Service
		Start     string
		End       string
		StartUnix int64
		EndUnix   int64
		Data      string
	}{serv, start.Format(utils.FlatpickrReadable), end.Format(utils.FlatpickrReadable), start.Unix(), end.Unix(), data.ToString()}

	executeResponse(w, r, "service.html", out, nil)
}

func servicesDeleteFailuresHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	service := core.SelectService(utils.StringInt(vars["id"]))
	service.DeleteFailures()
	executeResponse(w, r, "services.html", core.CoreApp.Services, "/services")
}

func checkinDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	checkin := core.SelectCheckinId(utils.StringInt(vars["id"]))
	service := core.SelectService(checkin.ServiceId)
	fmt.Println(checkin, service)
	checkin.Delete()
	executeResponse(w, r, "service.html", service, fmt.Sprintf("/service/%v", service.Id))
}

func checkinCreateHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	r.ParseForm()
	service := core.SelectService(utils.StringInt(vars["id"]))
	fmt.Println(service.Name)
	name := r.PostForm.Get("name")
	interval := utils.StringInt(r.PostForm.Get("interval"))
	grace := utils.StringInt(r.PostForm.Get("grace"))
	checkin := core.ReturnCheckin(&types.Checkin{
		Name:        name,
		ServiceId:   service.Id,
		Interval:    interval,
		GracePeriod: grace,
	})
	checkin.Create()
	executeResponse(w, r, "service.html", service, fmt.Sprintf("/service/%v", service.Id))
}

func checkinHitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	checkin := core.SelectCheckin(vars["id"])
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	checkinHit := core.ReturnCheckinHit(&types.CheckinHit{
		Checkin:   checkin.Id,
		From:      ip,
		CreatedAt: time.Now().UTC(),
	})
	if checkin.Last() == nil {
		checkin.Start()
		go checkin.Routine()
	}
	checkinHit.Create()
	w.Write([]byte("ok"))
	w.WriteHeader(http.StatusOK)
}
