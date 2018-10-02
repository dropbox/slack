package slack

import (
	"net/http"
	"testing"
	"github.com/stretchr/testify/require"
	"sync"
)

var registerHandlers sync.Once
var invalidChannel = "CXXXXXXXX"

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
		if r.FormValue("channel") == invalidChannel {
			writeSlackResponse(w, false, "channel_not_found")
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
	api := setUpClientForWorkspaceApp("bad-token", "")

	_, _, err := api.PostMessage("#test", "test", PostMessageParameters{})
	require.NotNil(t, err)
	require.Equal(t, "invalid_auth", err.Error())
}

func TestChatRefreshToken(t *testing.T) {
	registerHandlers.Do(mockHandleChatPostMessage)
	api := setUpClientForWorkspaceApp(testExpiredToken, testRefreshToken)

	_, _, err := api.PostMessage("#test", "test", PostMessageParameters{})
	require.Nil(t, err)
}

func TestPostMessageInvalidChannel(t *testing.T) {
	registerHandlers.Do(mockHandleChatPostMessage)
	api := setUpClientForWorkspaceApp(testValidToken, "")
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
