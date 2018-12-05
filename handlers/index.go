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

package handlers

import (
	"encoding/json"
	"github.com/hunterlong/statping/core"
	"github.com/hunterlong/statping/types"
	"github.com/hunterlong/statping/utils"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if core.Configs == nil {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}
	ExecuteResponse(w, r, "index.html", core.CoreApp, nil)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := map[string]interface{}{
		"services": len(core.Services()),
		"online":   core.Configs != nil,
	}
	json.NewEncoder(w).Encode(health)
}

func trayHandler(w http.ResponseWriter, r *http.Request) {
	ExecuteResponse(w, r, "tray.html", core.CoreApp, nil)
}

// DesktopInit will run the Statping server on a specific IP and port using SQLite database
func DesktopInit(ip string, port int) {
	var err error
	exists := utils.FileExists(utils.Directory + "/statup.db")
	if exists {
		core.Configs, err = core.LoadConfigFile(utils.Directory)
		if err != nil {
			utils.Log(3, err)
			return
		}
		err = core.Configs.Connect(false, utils.Directory)
		if err != nil {
			utils.Log(3, err)
			return
		}
		core.InitApp()
		RunHTTPServer(ip, port)
	}

	config := &core.DbConfig{
		DbConn:      "sqlite",
		Project:     "Statping",
		Description: "Statping running as an App!",
		Domain:      "http://localhost",
		Username:    "admin",
		Password:    "admin",
		Email:       "user@email.com",
		Error:       nil,
		Location:    utils.Directory,
	}

	config, err = config.Save()
	if err != nil {
		utils.Log(4, err)
	}

	config.DropDatabase()
	config.CreateDatabase()
	core.CoreApp = config.CreateCore()

	if err != nil {
		utils.Log(3, err)
		return
	}

	core.Configs, err = core.LoadConfigFile(utils.Directory)
	if err != nil {
		utils.Log(3, err)
		config.Error = err
		return
	}

	err = core.Configs.Connect(false, utils.Directory)
	if err != nil {
		utils.Log(3, err)
		core.DeleteConfig()
		config.Error = err
		return
	}

	admin := core.ReturnUser(&types.User{
		Username: config.Username,
		Password: config.Password,
		Email:    config.Email,
		Admin:    types.NewNullBool(true),
	})
	admin.Create()

	core.InsertSampleData()

	config.ApiKey = core.CoreApp.ApiKey
	config.ApiSecret = core.CoreApp.ApiSecret
	config.Update()

	core.InitApp()
	RunHTTPServer(ip, port)
}
