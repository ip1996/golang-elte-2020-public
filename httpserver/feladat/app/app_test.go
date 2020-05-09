package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Testing with testify and asserts
func TestRoutes(t *testing.T) {
	app := &App{}

	r := chi.NewRouter()
	app.Routes(r)

	require.Len(t, r.Routes(), 3)
	assert.True(t, hasRoutePatter(r.Routes(), "/"))
	assert.True(t, hasRoutePatter(r.Routes(), "/add"))
	assert.True(t, hasRoutePatter(r.Routes(), "/list"))
}

func hasRoutePatter(routes []chi.Route, pattern string) bool {
	for _, r := range routes {
		if r.Pattern == pattern {
			return true
		}
	}
	return false
}

// Testing with httptest's response recorder
func TestMainHandler(t *testing.T) {
	app := &App{} // No dependencies in the handler, so this should work

	rr := httptest.NewRecorder()
	app.handleMain(rr, injectNameIntoRequest(httptest.NewRequest(http.MethodGet, "/", nil)))

	require.Equal(t, http.StatusOK, rr.Code) // It's not changed, but test it anyway here for an example
	assert.Equal(t, "main app endpoint and stuff\n", rr.Body.String())
}

func TestListHandlerWithDBError(t *testing.T) {
	dbMock := &mockDatabase{}           // No New or any init step needed
	app := NewApp(zap.NewNop(), dbMock) // We'll need the DB as a dep, so use the New call

	// Create a fake request/response for the database
	// The mock will panic if there's and unexpected call: `assert: mock: I don't know what to return because the method call was unexpected.`
	dbMock.On("Select", mock.AnythingOfType("*[]app.message"), "SELECT id, name, message, created FROM message ORDER BY id DESC").
		Return(errors.New("mocked error")) // The "error" part of the response

	rr := httptest.NewRecorder()
	app.handleList(rr, injectNameIntoRequest(httptest.NewRequest(http.MethodGet, "/list", nil)))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"error":"mocked error","code":500}`+"\n", rr.Body.String())
}

func TestListHandlerWithEmptyResult(t *testing.T) {
	dbMock := &mockDatabase{}           // No New or any init step needed
	app := NewApp(zap.NewNop(), dbMock) // We'll need the DB as a dep, so use the New call

	// Create a fake request/response for the database
	// The mock will panic if there's and unexpected call: `assert: mock: I don't know what to return because the method call was unexpected.`
	dbMock.On("Select", mock.AnythingOfType("*[]app.message"), "SELECT id, name, message, created FROM message ORDER BY id DESC").
		Return(nil) // The "error" part of the response

	rr := httptest.NewRecorder()
	app.handleList(rr, injectNameIntoRequest(httptest.NewRequest(http.MethodGet, "/list", nil)))

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"messages":null}`+"\n", rr.Body.String())
}

func TestListHandlerWithResults(t *testing.T) {
	dbMock := &mockDatabase{}           // No New or any init step needed
	app := NewApp(zap.NewNop(), dbMock) // We'll need the DB as a dep, so use the New call

	// Create a fake request/response for the database
	// The mock will panic if there's and unexpected call: `assert: mock: I don't know what to return because the method call was unexpected.`
	dbMock.On("Select", mock.AnythingOfType("*[]app.message"), "SELECT id, name, message, created FROM message ORDER BY id DESC").
		// Some magic here: https://github.com/vektra/mockery#return-value-provider-functions
		Return(func(dest interface{}, query string, args ...interface{}) error {
			msgs, ok := dest.(*[]message)
			if !ok {
				return errors.New("invalid type for dest")
			}

			*msgs = []message{
				{
					ID:      1,
					Name:    "testName1",
					Message: "testMessage1",
					Created: time.Date(2020, 4, 27, 10, 10, 20, 0, time.UTC),
				},
				{
					ID:      2,
					Name:    "testName2",
					Message: "testMessage2",
					Created: time.Date(2020, 4, 27, 10, 11, 20, 0, time.UTC),
				},
			}
			return nil
		})

	rr := httptest.NewRecorder()
	app.handleList(rr, injectNameIntoRequest(httptest.NewRequest(http.MethodGet, "/list", nil)))

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"messages":[{"id":1,"user":"testName1","msg":"testMessage1","date":"2020-04-27T10:10:20Z"},{"id":2,"user":"testName2","msg":"testMessage2","date":"2020-04-27T10:11:20Z"}]}`+"\n", rr.Body.String())
}

func TestAddHandlerInvalidInput(t *testing.T) {
	dbMock := &mockDatabase{}
	app := NewApp(zap.NewNop(), dbMock)

	testTable := []struct {
		inputJSON      string
		expectedCode   int
		expectedOutput string
	}{
		{`invalid`, 400, `{"error":"invalid character 'i' looking for beginning of value","code":400}`},
		{`{}`, 400, `{"error":"message could not be empty","code":400}`},
		{`{"name":"a"}`, 400, `{"error":"message could not be empty","code":400}`},
	}

	rr := httptest.NewRecorder()
	for _, tt := range testTable {
		req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(tt.inputJSON))
		req = injectNameIntoRequest(req)
		app.handleAdd(rr, req)

		assert.Equal(t, tt.expectedCode, rr.Code)
		assert.Equal(t, tt.expectedOutput+"\n", rr.Body.String())
		rr.Body.Reset() // Don't forget this!
	}
}

//TODO (feladat): javitsd ki ezt a tesztet is!
func TestAddHandlerDBError(t *testing.T) {
	dbMock := &mockDatabase{}
	app := NewApp(zap.NewNop(), dbMock)

	dbMock.On("Exec", "INSERT INTO message (`name`, `message`) VALUES (?,?)", "dummyUser", "testMessage").
		Return(nil, errors.New("testError"))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name":"testName","msg":"testMessage"}`))
	req = injectNameIntoRequest(req)
	app.handleAdd(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"error":"testError","code":500}`+"\n", rr.Body.String())
}

// TODO(feladat): javitsd ki ezt a tesztet!
func TestAddHandlerOK(t *testing.T) {
	dbMock := &mockDatabase{}
	app := NewApp(zap.NewNop(), dbMock)

	dbMock.On("Exec", "INSERT INTO message (`name`, `message`) VALUES (?,?)", "dummyUser", "testMessage").
		Return(mockResult{
			id: 11,
		}, nil)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"msg":"testMessage"}`))
	req = injectNameIntoRequest(req)
	// <???>
	app.handleAdd(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"id":11}`+"\n", rr.Body.String())
}

type mockResult struct {
	id int64
}

func (mr mockResult) LastInsertId() (int64, error) {
	return mr.id, nil
}

func (mockResult) RowsAffected() (int64, error) {
	return 0, nil
}

func TestSendError(t *testing.T) {
	mel := &mockErrorLogger{}
	rr := httptest.NewRecorder()
	sendError(mel, rr, errors.New("testError"), 444)

	assert.Equal(t, "error on backend", mel.msg)
	assert.Equal(t, `{"error":"testError","code":444}`+"\n", rr.Body.String())
}

type mockErrorLogger struct {
	msg    string
	fields []zap.Field
}

func (mel *mockErrorLogger) Error(msg string, fields ...zap.Field) {
	mel.msg = msg
	mel.fields = fields
}

func injectNameIntoRequest(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), nameContextKey, "dummyUser"))
}
