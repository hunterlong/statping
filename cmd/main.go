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

package main

import (
	"flag"
	"fmt"
	"github.com/hunterlong/statup/core"
	"github.com/hunterlong/statup/handlers"
	_ "github.com/hunterlong/statup/notifiers"
	"github.com/hunterlong/statup/plugin"
	"github.com/hunterlong/statup/source"
	"github.com/hunterlong/statup/utils"
	"github.com/joho/godotenv"
	"os"
)

var (
	// VERSION stores the current version of Statup
	VERSION string
	// COMMIT stores the git commit hash for this version of Statup
	COMMIT      string
	ipAddress   string
	UsingDotEnv bool
	port        int
)

func init() {
	core.VERSION = VERSION
}

// parseFlags will parse the application flags
// -ip = 0.0.0.0 IP address for outgoing HTTP server
// -port = 8080 Port number for outgoing HTTP server
func parseFlags() {
	ip := flag.String("ip", "0.0.0.0", "IP address to run the Statup HTTP server")
	p := flag.Int("port", 8080, "Port to run the HTTP server")
	flag.Parse()
	ipAddress = *ip
	port = *p
}

// main will run the Statup application
func main() {
	var err error
	parseFlags()
	LoadDotEnvs()
	source.Assets()
	utils.InitLogs()
	args := flag.Args()

	if len(args) >= 1 {
		err := CatchCLI(args)
		if err != nil {
			if err.Error() == "end" {
				os.Exit(0)
			}
			fmt.Println(err)
			os.Exit(1)
		}
	}
	utils.Log(1, fmt.Sprintf("Starting Statup v%v", VERSION))
	core.Configs, err = core.LoadConfig(utils.Directory)
	if err != nil {
		utils.Log(3, err)
		core.SetupMode = true
		fmt.Println(handlers.RunHTTPServer(ipAddress, port))
		os.Exit(1)
	}
	defer core.CloseDB()
	mainProcess()
}

// LoadDotEnvs attempts to load database configs from a '.env' file in root directory
func LoadDotEnvs() error {
	err := godotenv.Load()
	if err == nil {
		utils.Log(1, "Environment file '.env' Loaded")
		UsingDotEnv = true
	}
	return err
}

// mainProcess will initialize the Statup application and run the HTTP server
func mainProcess() {
	dir := utils.Directory
	var err error
	core.Configs, err = core.LoadConfig(dir)
	if err != nil {
		utils.Log(4, fmt.Sprintf("could not load config.yml %v", err))
	}
	err = core.Configs.Connect(false, dir)
	if err != nil {
		utils.Log(4, fmt.Sprintf("could not connect to database: %v", err))
	}
	core.Configs.MigrateDatabase()
	core.InitApp()
	if !core.SetupMode {
		plugin.LoadPlugins()
		fmt.Println(handlers.RunHTTPServer(ipAddress, port))
		os.Exit(1)
	}
}

func ForEachPlugin() {
	if len(core.CoreApp.Plugins) > 0 {
		//for _, p := range core.Plugins {
		//	p.OnShutdown()
		//}
	}
}
