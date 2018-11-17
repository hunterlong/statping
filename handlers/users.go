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
	"errors"
	"github.com/gorilla/mux"
	"github.com/hunterlong/statup/core"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"net/http"
	"strconv"
)

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	users, _ := core.SelectAllUsers()
	executeResponse(w, r, "users.html", users, nil)
}

func usersEditHandler(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	user, _ := core.SelectUser(int64(id))
	executeResponse(w, r, "user.html", user, nil)
}

func apiUserHandler(w http.ResponseWriter, r *http.Request) {
	if !isAPIAuthorized(r) {
		sendUnauthorizedJson(w, r)
		return
	}
	vars := mux.Vars(r)
	user, err := core.SelectUser(utils.StringInt(vars["id"]))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func apiUserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !isAPIAuthorized(r) {
		sendUnauthorizedJson(w, r)
		return
	}
	vars := mux.Vars(r)
	user, err := core.SelectUser(utils.StringInt(vars["id"]))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	var updateUser *types.User
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&updateUser)
	updateUser.Id = user.Id
	user = core.ReturnUser(updateUser)
	err = user.Update()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func apiUserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAPIAuthorized(r) {
		sendUnauthorizedJson(w, r)
		return
	}
	vars := mux.Vars(r)
	users := core.CountUsers()
	if users == 1 {
		sendErrorJson(errors.New("cannot delete the last user"), w, r)
		return
	}
	user, err := core.SelectUser(utils.StringInt(vars["id"]))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	err = user.Delete()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	output := apiResponse{
		Object: "user",
		Method: "delete",
		Id:     user.Id,
		Status: "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func apiAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !isAPIAuthorized(r) {
		sendUnauthorizedJson(w, r)
		return
	}
	users, err := core.SelectAllUsers()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func apiCreateUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !isAPIAuthorized(r) {
		sendUnauthorizedJson(w, r)
		return
	}
	var user *types.User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	newUser := core.ReturnUser(user)
	uId, err := newUser.Create()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	output := apiResponse{
		Object: "user",
		Method: "create",
		Id:     uId,
		Status: "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
