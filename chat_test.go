package slack

import (
	"encoding/json"
	"net/http"
	"testing"
	"github.com/stretchr/testify/require"
)

var registerHandlers sync.Once

func mockHandleChatPostMessage() {
	http.HandleFunc("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		if token == "" {
			writeSlackResponse(w, false, "not_authed")
			return
		}
		if token != testValidToken {
			writeSlackResponse(w, false, "invalid_auth")
			return
		}

		writeSlackResponse(w, true, "")
	})
}

func TestChatValidToken(t *testing.T) {
	registerHandlers.Do(mockHandleChatPostMessage)
	api := setUpClientForWorkspaceApp(testValidToken, "")

	_, _, err := api.PostMessage("#test", "test", PostMessageParameters{})
	require.Nil(t, err)
}

func TestChatInvalidToken(t *testing.T) {
	registerHandlers.Do(mockHandleChatPostMessage)
	api := setUpClientForWorkspaceApp("bad-token", "")

	_, _, err := api.PostMessage("#test", "test", PostMessageParameters{})
	require.NotNil(t, err)
}

func TestChatRefreshToken(t *testing.T) {
	registerHandlers.Do(mockHandleChatPostMessage)
	api := setUpClientForWorkspaceApp(testExpiredToken, testRefreshToken)

	_, _, err := api.PostMessage("#test", "test", PostMessageParameters{})
	require.Nil(t, err)
}


func postMessageInvalidChannelHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(chatResponseFull{
		SlackResponse: SlackResponse{Ok: false, Error: "channel_not_found"},
	})
	rw.Write(response)
}

func TestPostMessageInvalidChannel(t *testing.T) {
	http.HandleFunc("/chat.postMessage", postMessageInvalidChannelHandler)
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	_, _, err := api.PostMessage("CXXXXXXXX", "hello", PostMessageParameters{})
	if err == nil {
		t.Errorf("Expected error: %s; instead succeeded", "channel_not_found")
		return
	}

	if err.Error() != "channel_not_found" {
		t.Errorf("Expected error: %s; received: %s", "channel_not_found", err)
		return
	}
}
