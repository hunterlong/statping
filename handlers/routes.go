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
	"fmt"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hunterlong/statping/core"
	"github.com/hunterlong/statping/handlers/graphql"
	"github.com/hunterlong/statping/source"
	"github.com/hunterlong/statping/utils"
	"net/http"
	"os"
)

var (
	router *mux.Router
	log    = utils.Log.WithField("type", "handlers")
)

// Router returns all of the routes used in Statping.
// Server will use static assets if the 'assets' directory is found in the root directory.
func Router() *mux.Router {
	dir := utils.Directory
	CacheStorage = NewStorage()
	r := mux.NewRouter().StrictSlash(true)

	if os.Getenv("AUTH_USERNAME") != "" && os.Getenv("AUTH_PASSWORD") != "" {
		authUser = os.Getenv("AUTH_USERNAME")
		authPass = os.Getenv("AUTH_PASSWORD")
		r.Use(basicAuthHandler)
	}

	if os.Getenv("BASE_PATH") != "" {
		basePath = "/" + os.Getenv("BASE_PATH") + "/"
		r = r.PathPrefix("/" + os.Getenv("BASE_PATH")).Subrouter()
		r.Handle("", http.HandlerFunc(indexHandler))
	} else {
		r.Handle("/", http.HandlerFunc(indexHandler))
	}

	r.Use(sendLog)
	if source.UsingAssets(dir) {
		indexHandler := http.FileServer(http.Dir(dir + "/assets/"))
		r.PathPrefix("/css/").Handler(http.StripPrefix(basePath+"css/", http.FileServer(http.Dir(dir+"/assets/css"))))
		r.PathPrefix("/font/").Handler(http.StripPrefix(basePath+"font/", http.FileServer(http.Dir(dir+"/assets/font"))))
		r.PathPrefix("/js/").Handler(http.StripPrefix(basePath+"js/", http.FileServer(http.Dir(dir+"/assets/js"))))
		r.PathPrefix("/robots.txt").Handler(http.StripPrefix(basePath, indexHandler))
		r.PathPrefix("/favicon.ico").Handler(http.StripPrefix(basePath, indexHandler))
		r.PathPrefix("/banner.png").Handler(http.StripPrefix(basePath, indexHandler))
	} else {
		r.PathPrefix("/css/").Handler(http.StripPrefix(basePath+"css/", http.FileServer(source.CssBox.HTTPBox())))
		r.PathPrefix("/font/").Handler(http.StripPrefix(basePath+"font/", http.FileServer(source.FontBox.HTTPBox())))
		r.PathPrefix("/js/").Handler(http.StripPrefix(basePath+"js/", http.FileServer(source.JsBox.HTTPBox())))
		r.PathPrefix("/robots.txt").Handler(http.StripPrefix(basePath, http.FileServer(source.TmplBox.HTTPBox())))
		r.PathPrefix("/favicon.ico").Handler(http.StripPrefix(basePath, http.FileServer(source.TmplBox.HTTPBox())))
		r.PathPrefix("/banner.png").Handler(http.StripPrefix(basePath, http.FileServer(source.TmplBox.HTTPBox())))
	}
	r.Handle("/charts.js", http.HandlerFunc(renderServiceChartsHandler))
	r.Handle("/setup", http.HandlerFunc(setupHandler)).Methods("GET")
	r.Handle("/setup", http.HandlerFunc(processSetupHandler)).Methods("POST")
	r.Handle("/dashboard", http.HandlerFunc(dashboardHandler)).Methods("GET")
	r.Handle("/dashboard", http.HandlerFunc(loginHandler)).Methods("POST")
	r.Handle("/logout", http.HandlerFunc(logoutHandler))
	r.Handle("/plugins/download/{name}", authenticated(pluginsDownloadHandler, true))
	r.Handle("/plugins/{name}/save", authenticated(pluginSavedHandler, true)).Methods("POST")
	r.Handle("/help", authenticated(helpHandler, true))
	r.Handle("/logs", authenticated(logsHandler, true))
	r.Handle("/logs/line", readOnly(logsLineHandler, true))

	// GRAPHQL Route
	r.Handle("/graphql", authenticated(handler.GraphQL(graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{}})), true))

	// USER Routes
	r.Handle("/users", readOnly(usersHandler, true)).Methods("GET")
	r.Handle("/user/{id}", authenticated(usersEditHandler, true)).Methods("GET")

	// MESSAGES Routes
	r.Handle("/messages", authenticated(messagesHandler, true)).Methods("GET")
	r.Handle("/message/{id}", authenticated(viewMessageHandler, true)).Methods("GET")

	// SETTINGS Routes
	r.Handle("/settings", authenticated(settingsHandler, true)).Methods("GET")
	r.Handle("/settings", authenticated(saveSettingsHandler, true)).Methods("POST")
	r.Handle("/settings/css", authenticated(saveSASSHandler, true)).Methods("POST")
	r.Handle("/settings/build", authenticated(saveAssetsHandler, true)).Methods("GET")
	r.Handle("/settings/delete_assets", authenticated(deleteAssetsHandler, true)).Methods("GET")
	r.Handle("/settings/export", authenticated(exportHandler, true)).Methods("GET")
	r.Handle("/settings/bulk_import", authenticated(bulkImportHandler, true)).Methods("POST")
	r.Handle("/settings/integrator/{name}", authenticated(integratorHandler, true)).Methods("POST")

	// SERVICE Routes
	r.Handle("/services", authenticated(servicesHandler, true)).Methods("GET")
	r.Handle("/service/create", authenticated(createServiceHandler, true)).Methods("GET")
	r.Handle("/service/{id}", http.HandlerFunc(servicesViewHandler)).Methods("GET")
	r.Handle("/service/{id}/edit", authenticated(servicesViewHandler, true)).Methods("GET")
	r.Handle("/service/{id}/delete_failures", authenticated(servicesDeleteFailuresHandler, true)).Methods("POST")

	r.Handle("/group/{id}", http.HandlerFunc(groupViewHandler)).Methods("GET")

	// API Routes
	r.Handle("/api", authenticated(apiIndexHandler, false))
	r.Handle("/api/renew", authenticated(apiRenewHandler, false))
	r.Handle("/api/clear_cache", authenticated(apiClearCacheHandler, false))

	r.Handle("/api/integrations", authenticated(apiAllIntegrationsHandler, false)).Methods("GET")
	r.Handle("/api/integrations/{name}", authenticated(apiIntegrationHandler, false)).Methods("GET")
	r.Handle("/api/integrations/{name}", authenticated(apiIntegrationHandler, false)).Methods("POST")

	// API GROUPS Routes
	r.Handle("/api/groups", readOnly(apiAllGroupHandler, false)).Methods("GET")
	r.Handle("/api/groups", authenticated(apiCreateGroupHandler, false)).Methods("POST")
	r.Handle("/api/groups/{id}", readOnly(apiGroupHandler, false)).Methods("GET")
	r.Handle("/api/groups/{id}", authenticated(apiGroupUpdateHandler, false)).Methods("POST")
	r.Handle("/api/groups/{id}", authenticated(apiGroupDeleteHandler, false)).Methods("DELETE")
	r.Handle("/api/reorder/groups", authenticated(apiGroupReorderHandler, false)).Methods("POST")

	// API SERVICE Routes
	r.Handle("/api/services", readOnly(apiAllServicesHandler, false)).Methods("GET")
	r.Handle("/api/services", authenticated(apiCreateServiceHandler, false)).Methods("POST")
	r.Handle("/api/services/{id}", readOnly(apiServiceHandler, false)).Methods("GET")
	r.Handle("/api/reorder/services", authenticated(reorderServiceHandler, false)).Methods("POST")
	r.Handle("/api/services/{id}/running", authenticated(apiServiceRunningHandler, false)).Methods("POST")
	r.Handle("/api/services/{id}/data", cached("30s", "application/json", apiServiceDataHandler)).Methods("GET")
	r.Handle("/api/services/{id}/ping", cached("30s", "application/json", apiServicePingDataHandler)).Methods("GET")
	r.Handle("/api/services/{id}/heatmap", cached("30s", "application/json", apiServiceHeatmapHandler)).Methods("GET")
	r.Handle("/api/services/{id}", authenticated(apiServiceUpdateHandler, false)).Methods("POST")
	r.Handle("/api/services/{id}", authenticated(apiServiceDeleteHandler, false)).Methods("DELETE")
	r.Handle("/api/services/{id}/failures", authenticated(apiServiceFailuresHandler, false)).Methods("GET")
	r.Handle("/api/services/{id}/failures", authenticated(servicesDeleteFailuresHandler, false)).Methods("DELETE")
	r.Handle("/api/services/{id}/hits", authenticated(apiServiceHitsHandler, false)).Methods("GET")

	// API INCIDENTS Routes
	r.Handle("/api/incidents", readOnly(apiAllIncidentsHandler, false)).Methods("GET")

	// API USER Routes
	r.Handle("/api/users", authenticated(apiAllUsersHandler, false)).Methods("GET")
	r.Handle("/api/users", authenticated(apiCreateUsersHandler, false)).Methods("POST")
	r.Handle("/api/users/{id}", authenticated(apiUserHandler, false)).Methods("GET")
	r.Handle("/api/users/{id}", authenticated(apiUserUpdateHandler, false)).Methods("POST")
	r.Handle("/api/users/{id}", authenticated(apiUserDeleteHandler, false)).Methods("DELETE")

	// API NOTIFIER Routes
	r.Handle("/api/notifiers", authenticated(apiNotifiersHandler, false)).Methods("GET")
	r.Handle("/api/notifier/{notifier}", authenticated(apiNotifierGetHandler, false)).Methods("GET")
	r.Handle("/api/notifier/{notifier}", authenticated(apiNotifierUpdateHandler, false)).Methods("POST")
	r.Handle("/api/notifier/{method}/test", authenticated(testNotificationHandler, false)).Methods("POST")

	// API MESSAGES Routes
	r.Handle("/api/messages", readOnly(apiAllMessagesHandler, false)).Methods("GET")
	r.Handle("/api/messages", authenticated(apiMessageCreateHandler, false)).Methods("POST")
	r.Handle("/api/messages/{id}", readOnly(apiMessageGetHandler, false)).Methods("GET")
	r.Handle("/api/messages/{id}", authenticated(apiMessageUpdateHandler, false)).Methods("POST")
	r.Handle("/api/messages/{id}", authenticated(apiMessageDeleteHandler, false)).Methods("DELETE")

	// API CHECKIN Routes
	r.Handle("/api/checkins", authenticated(apiAllCheckinsHandler, false)).Methods("GET")
	r.Handle("/api/checkin/{api}", authenticated(apiCheckinHandler, false)).Methods("GET")
	r.Handle("/api/checkin", authenticated(checkinCreateHandler, false)).Methods("POST")
	r.Handle("/api/checkin/{api}", authenticated(checkinDeleteHandler, false)).Methods("DELETE")
	r.Handle("/checkin/{api}", http.HandlerFunc(checkinHitHandler))

	// Static Files Routes
	r.PathPrefix("/files/postman.json").Handler(http.StripPrefix("/files/", http.FileServer(source.TmplBox.HTTPBox())))
	r.PathPrefix("/files/swagger.json").Handler(http.StripPrefix("/files/", http.FileServer(source.TmplBox.HTTPBox())))
	r.PathPrefix("/files/grafana.json").Handler(http.StripPrefix("/files/", http.FileServer(source.TmplBox.HTTPBox())))

	// API Generic Routes
	r.Handle("/metrics", readOnly(prometheusHandler, false))
	r.Handle("/health", http.HandlerFunc(healthCheckHandler))
	r.Handle("/.well-known/", http.StripPrefix("/.well-known/", http.FileServer(http.Dir(dir+"/.well-known"))))

	r.NotFoundHandler = http.HandlerFunc(error404Handler)
	return r
}

func resetRouter() {
	router = Router()
	httpServer.Handler = router
}

func resetCookies() {
	if core.CoreApp != nil {
		cookie := fmt.Sprintf("%v_%v", core.CoreApp.ApiSecret, utils.Now().Nanosecond())
		sessionStore = sessions.NewCookieStore([]byte(cookie))
	} else {
		sessionStore = sessions.NewCookieStore([]byte("secretinfo"))
	}
}
