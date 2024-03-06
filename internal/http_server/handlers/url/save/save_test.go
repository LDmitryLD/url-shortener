package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LDmitryLD/url-shortener/internal/http_server/handlers/url/save/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSaveHandler(t *testing.T) {
	logger := zap.NewExample()

	cases := []struct {
		name       string
		alias      string
		url        string
		respoError string
		mockError  error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
		},
		{
			name:       "Empty URL",
			url:        "",
			alias:      "some_alias",
			respoError: "field URL is a required field",
		},
		{
			name:       "Invalid URL",
			url:        "some invalid URL",
			alias:      "some_alias",
			respoError: "field URL is not a valid URL",
		},
		{
			name:       "SaveURL Error",
			alias:      "test_alias",
			url:        "https://google.com",
			respoError: "failed to add url",
			mockError:  errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaveMock := mocks.NewURLSaver(t)

			if tc.respoError == "" || tc.mockError != nil {
				urlSaveMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			handler := New(logger, urlSaveMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respoError, resp.Error)
		})
	}
}
