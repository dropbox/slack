package slack

import (
	"log"
	"net/http/httptest"
	"sync"
	"net/http"
	"bytes"
	"encoding/json"
)

const (
	testValidToken = "testing-token"
	testClientId     = "client-id"
	testClientSecret = "client-secret"
	testExpiredToken = "expired-token"
	testRefreshToken = "refresh-token"
)

var (
	serverAddr string
	once       sync.Once
)

func setUpClientForWorkspaceApp(authToken string, refreshToken string) *Client  {
	once.Do(startServer)
	refreshConfig := RefreshTokenConfig{
		RefreshToken: refreshToken,
		ClientId: testClientId,
		ClientSecret: testClientSecret,
	}
	return NewWithRefreshToken(authToken, refreshConfig)
}

func startServer() {
	http.HandleFunc("/oauth.access", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("client_id") != testClientId {
			writeSlackResponse(w, false, "invalid_client_id")
			return
		}
		if r.FormValue("client_secret") != testClientSecret {
			writeSlackResponse(w, false, "invalid_client_secret")
			return
		}
		formRefreshToken := r.FormValue("refresh_token")
		formGrantType := r.FormValue("grant_type")
		if formGrantType == "refresh_token" && formRefreshToken == testRefreshToken {
			resp := OAuthResponse{
				AccessToken: testValidToken,
				SlackResponse: SlackResponse{
					Ok: true,
				},
			}
			writeResponse(w, resp)
			return
		}

		response := []byte(`{"ok": false,"error":"unknown"}`)
		w.Write(response)
	})

	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test WebSocket server listening on ", serverAddr)
	SLACK_API = "http://" + serverAddr + "/"
}

func writeSlackResponse(w http.ResponseWriter, ok bool, err string){
	resp := SlackResponse {
		Ok: ok,
		Error: err,
	}
	writeResponse(w, resp)
}

func writeResponse(w http.ResponseWriter, resp interface{}) {
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(resp)
	w.Write(b.Bytes())
}

