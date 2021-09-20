package redirect

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var istioMeshId = os.Getenv("ISTIO_MESH_ID")
var listBaseURL = "https://console.cloud.google.com/traces/list?orgonly=true&supportedpurview=organizationId&project=" + os.Getenv("GOOGLE_CLOUD_PROJECT")
func handleHome(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, listBaseURL, http.StatusSeeOther)
}

// GET /search?service=istio-ingressgateway&start=1632036540916000&limit=100
func handleSearch(w http.ResponseWriter, req *http.Request) {
	searchURL := listBaseURL

	// collect filter tags
	tags := map[string]string {}
	query := req.URL.Query()
	if query.Has("tags") {
		if err := json.Unmarshal([]byte(query.Get("tags")), &tags); err != nil {
			log.Printf("Failed to unmarshal tags query: %v", query.Get("tags"))
		}
	}
	if len(istioMeshId) > 0 {
		tags["istio.mesh_id"] = istioMeshId
	}
	if query.Has("service") {
		// eg. serviceName = istio-ingressgateway.istio-system
		serviceNameTokens := strings.Split(query.Get("service"), ".")
		tags["istio.canonical_service"] = serviceNameTokens[0]
		if len(serviceNameTokens) > 1 {
			tags["istio.namespace"] = serviceNameTokens[1]
		}
	}

	// calculate duration
	interval := "P1D"
	if query.Has("start") {
		startTimestampMicro, err := strconv.ParseInt(query.Get("start"), 10, 64)
		if err != nil {
			log.Printf("Failed to unmarshal start query: %v", query.Get("start"))
		} else {
			duration := time.Now().Sub(time.UnixMicro(startTimestampMicro))
			log.Printf("duration is.. %v", duration)
			if duration <= 1 * time.Hour {
				interval = "PT1H"
			} else if duration <= 6 * time.Hour {
				interval = "PT6H"
			} else if duration <= 12 * time.Hour {
				interval = "PT12H"
			} else if duration <= 24 * time.Hour {
				interval = "P1D"
			} else if duration <= 2 * 24 * time.Hour {
				interval = "P2D"
			} else if duration <= 4 * 24 * time.Hour {
				interval = "P4D"
			} else if duration <= 7 * 24 * time.Hour {
				interval = "P7D"
			} else if duration <= 14 * 24 * time.Hour {
				interval = "P14D"
			} else {
				interval = "P30D"
			}
		}

		if query.Has("end") {
			endTimestampMicro, err := strconv.ParseInt(query.Get("end"), 10, 64)
			if err != nil {
				log.Printf("Failed to unmarshal start query: %v", query.Get("start"))
			} else {
				searchURL += "&start=" + strconv.FormatInt(startTimestampMicro/1000, 10) + "&end=" + strconv.FormatInt(endTimestampMicro/1000, 10)
			}
		}
	}

	searchURL += `&pageState=("traceFilter":("chips":"`
	searchURL += "%255B"
	i := 0
	for k, v := range tags {
		searchURL += `%257B_22k_22_3A_22LABEL_3A`+k+`_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22`+v+`_5C_22_22_2C_22s_22_3Atrue_2C_22i_22_3A_22`+k+`_22%257D`
		i++
		if i < len(tags) {
			searchURL += "_2C"
		}
	}
	searchURL += "%255D"
	searchURL += `"),"traceIntervalPicker":("groupValue":"` + interval + `","customValue":null))`

	http.Redirect(w, req, searchURL, http.StatusSeeOther)
}

func ServeHTTPRedirectServer(host string) error {
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/", handleHome)
	return http.ListenAndServe(host, nil)
}