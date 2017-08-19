package auth

import (
	"io/ioutil"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

// Forward the authentication to a external server
func Forward(forward *types.Forward, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	client := http.Client{}

	forwardReq, err := http.NewRequest("GET", forward.Address, nil)
	if err != nil {
		log.Debugf("Error calling %s. Cause %s", forward.Address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	forwardReq.Header = r.Header

	forwardResponse, forwardErr := client.Do(forwardReq)
	if forwardErr != nil {
		log.Debugf("Error calling %s. Cause: %s", forward.Address, forwardErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer forwardResponse.Body.Close()
	body, readError := ioutil.ReadAll(forwardResponse.Body)
	if readError != nil {
		log.Debugf("Error reading body %s. Cause: %s", forward.Address, readError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if forwardResponse.StatusCode < 200 || forwardResponse.StatusCode >= 300 {
		log.Debugf("Remote error %s. StatusCode: %d", forward.Address, forwardResponse.StatusCode)
		w.WriteHeader(forwardResponse.StatusCode)
		w.Write(body)
		return
	}

	r.RequestURI = r.URL.RequestURI()
	next(w, r)
}
