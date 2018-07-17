package handlers

import (
	"fmt"
	"github.com/hunterlong/statup/core"
	"github.com/hunterlong/statup/utils"
	"net/http"
	"strings"
)

//
//   Use your Statup Secret API Key for Authentication
//
//		scrape_configs:
//		  - job_name: 'statup'
//			bearer_token: MY API SECRET HERE
//			static_configs:
//			  - targets: ['statup:8080']
//

func PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	utils.Log(1, fmt.Sprintf("Prometheus /metrics Request From IP: %v\n", r.RemoteAddr))
	if !isAuthorized(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	metrics := []string{}
	system := fmt.Sprintf("statup_total_failures %v\n", core.CountFailures())
	system += fmt.Sprintf("statup_total_services %v", len(core.CoreApp.Services))
	metrics = append(metrics, system)
	for _, ser := range core.CoreApp.Services {
		v := ser.ToService()
		online := 1
		if !v.Online {
			online = 0
		}
		met := fmt.Sprintf("statup_service_failures{id=\"%v\" name=\"%v\"} %v\n", v.Id, v.Name, len(v.Failures))
		met += fmt.Sprintf("statup_service_latency{id=\"%v\" name=\"%v\"} %0.0f\n", v.Id, v.Name, (v.Latency * 100))
		met += fmt.Sprintf("statup_service_online{id=\"%v\" name=\"%v\"} %v\n", v.Id, v.Name, online)
		met += fmt.Sprintf("statup_service_status_code{id=\"%v\" name=\"%v\"} %v\n", v.Id, v.Name, v.LastStatusCode)
		met += fmt.Sprintf("statup_service_response_length{id=\"%v\" name=\"%v\"} %v", v.Id, v.Name, len([]byte(v.LastResponse)))
		metrics = append(metrics, met)
	}
	output := strings.Join(metrics, "\n")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

func isAuthorized(r *http.Request) bool {
	var token string
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token == core.CoreApp.ApiSecret {
		return true
	}
	return false
}
