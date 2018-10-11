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
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"io/ioutil"
	"os"
)

// ErrorResponse is used for HTTP errors to show to user
type ErrorResponse struct {
	Error string
}

// LoadConfig will attempt to load the 'config.yml' file in a specific directory
func LoadConfig(directory string) (*DbConfig, error) {
	var configs *DbConfig
	if os.Getenv("DB_CONN") != "" {
		utils.Log(1, "DB_CONN environment variable was found, waiting for database...")
		return LoadUsingEnv()
	}
	file, err := ioutil.ReadFile(directory + "/config.yml")
	if err != nil {
		return nil, errors.New("config.yml file not found at " + directory + "/config.yml - starting in setup mode")
	}
	err = yaml.Unmarshal(file, &configs)
	if err != nil {
		return nil, err
	}
	Configs = configs
	return Configs, err
}

// LoadUsingEnv will attempt to load database configs based on environment variables. If DB_CONN is set if will force this function.
func LoadUsingEnv() (*DbConfig, error) {
	Configs = new(DbConfig)
	if os.Getenv("DB_CONN") == "" {
		return nil, errors.New("Missing DB_CONN environment variable")
	}
	if os.Getenv("DB_HOST") == "" {
		return nil, errors.New("Missing DB_HOST environment variable")
	}
	if os.Getenv("DB_USER") == "" {
		return nil, errors.New("Missing DB_USER environment variable")
	}
	if os.Getenv("DB_PASS") == "" {
		return nil, errors.New("Missing DB_PASS environment variable")
	}
	if os.Getenv("DB_DATABASE") == "" {
		return nil, errors.New("Missing DB_DATABASE environment variable")
	}
	Configs.DbConn = os.Getenv("DB_CONN")
	Configs.DbHost = os.Getenv("DB_HOST")
	Configs.DbPort = int(utils.StringInt(os.Getenv("DB_PORT")))
	Configs.DbUser = os.Getenv("DB_USER")
	Configs.DbPass = os.Getenv("DB_PASS")
	Configs.DbData = os.Getenv("DB_DATABASE")
	CoreApp.DbConnection = os.Getenv("DB_CONN")
	CoreApp.Name = os.Getenv("NAME")
	CoreApp.Domain = os.Getenv("DOMAIN")
	if os.Getenv("USE_CDN") == "true" {
		CoreApp.UseCdn = true
	}

	dbConfig := &DbConfig{
		DbConn:      os.Getenv("DB_CONN"),
		DbHost:      os.Getenv("DB_HOST"),
		DbUser:      os.Getenv("DB_USER"),
		DbPass:      os.Getenv("DB_PASS"),
		DbData:      os.Getenv("DB_DATABASE"),
		DbPort:      5432,
		Project:     "Statup - " + os.Getenv("NAME"),
		Description: "New Statup Installation",
		Domain:      os.Getenv("DOMAIN"),
		Username:    "admin",
		Password:    "admin",
		Email:       "info@localhost.com",
	}

	err := dbConfig.Connect(true, utils.Directory)
	if err != nil {
		utils.Log(4, err)
		return nil, err
	}

	exists := DbSession.HasTable("core")
	if !exists {
		utils.Log(1, fmt.Sprintf("Core database does not exist, creating now!"))
		dbConfig.DropDatabase()
		dbConfig.CreateDatabase()

		CoreApp, err = dbConfig.InsertCore()
		if err != nil {
			utils.Log(3, err)
		}

		admin := &types.User{
			Username: "admin",
			Password: "admin",
			Email:    "info@admin.com",
			Admin:    true,
		}
		admin.Create()

		InsertSampleData()

		return Configs, err

	}

	return Configs, nil
}

// DeleteConfig will delete the 'config.yml' file
func DeleteConfig() {
	err := os.Remove(utils.Directory + "/config.yml")
	if err != nil {
		utils.Log(3, err)
	}
}
