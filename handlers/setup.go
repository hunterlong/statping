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
	"github.com/hunterlong/statup/core"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"net/http"
	"os"
	"time"
)

func setupHandler(w http.ResponseWriter, r *http.Request) {
	if core.CoreApp.Services != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var data interface{}
	if os.Getenv("DB_CONN") != "" {
		data, _ = core.LoadUsingEnv()
	}
	w.WriteHeader(http.StatusOK)
	executeResponse(w, r, "setup.html", data, nil)
}

func processSetupHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if core.Services() != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	dbHost := r.PostForm.Get("db_host")
	dbUser := r.PostForm.Get("db_user")
	dbPass := r.PostForm.Get("db_password")
	dbDatabase := r.PostForm.Get("db_database")
	dbConn := r.PostForm.Get("db_connection")
	dbPort := utils.StringInt(r.PostForm.Get("db_port"))
	project := r.PostForm.Get("project")
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	//sample := r.PostForm.Get("sample_data")
	description := r.PostForm.Get("description")
	domain := r.PostForm.Get("domain")
	email := r.PostForm.Get("email")
	dir := utils.Directory

	config := &core.DbConfig{
		DbConn:      dbConn,
		DbHost:      dbHost,
		DbUser:      dbUser,
		DbPass:      dbPass,
		DbData:      dbDatabase,
		DbPort:      dbPort,
		Project:     project,
		Description: description,
		Domain:      domain,
		Username:    username,
		Password:    password,
		Email:       email,
		Error:       nil,
		Location:    utils.Directory,
	}

	if core.Configs, err = config.Save(); err != nil {
		utils.Log(4, err)
		config.Error = err
		setupResponseError(w, r, config)
		return
	}

	if core.Configs, err = core.LoadConfigFile(dir); err != nil {
		utils.Log(3, err)
		config.Error = err
		setupResponseError(w, r, config)
		return
	}

	if err = core.Configs.Connect(false, dir); err != nil {
		utils.Log(4, err)
		core.DeleteConfig()
		config.Error = err
		setupResponseError(w, r, config)
		return
	}

	config.DropDatabase()
	config.CreateDatabase()

	core.CoreApp, err = config.InsertCore()
	if err != nil {
		utils.Log(4, err)
		config.Error = err
		setupResponseError(w, r, config)
		return
	}

	admin := core.ReturnUser(&types.User{
		Username: config.Username,
		Password: config.Password,
		Email:    config.Email,
		Admin:    types.NewNullBool(true),
	})
	admin.Create()

	core.SampleData()
	core.InitApp()
	CacheStorage.Delete("/")
	resetCookies()
	time.Sleep(2 * time.Second)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func setupResponseError(w http.ResponseWriter, r *http.Request, a interface{}) {
	executeResponse(w, r, "setup.html", a, nil)
}
