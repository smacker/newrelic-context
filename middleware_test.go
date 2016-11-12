package nrcontext

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		tnx := GetTnxFromContext(r.Context())
		if tnx == nil {
			t.Fatal("can't get tnx from context")
		}

		w.Write([]byte("Test response"))
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	app := &NewrelicAppMock{}
	nr := &NewRelicMiddleware{
		app:      app,
		nameFunc: func(r *http.Request) string { return r.URL.Path },
	}

	handler := nr.Handler(http.HandlerFunc(testHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("status code is wrong o_O")
	}
	if app.Tnx.GetName() != "/foo" {
		t.Errorf("transaction name is wrong: %v", app.Tnx.GetName())
	}
	if app.Tnx.WasEnded != true {
		t.Error("transaction didn't finish")
	}

	nr.SetTxnNameFunc(func(r *http.Request) string {
		return "test"
	})

	handler.ServeHTTP(w, req)
	if app.Tnx.GetName() != "test" {
		t.Errorf("transaction name is wrong: %v", app.Tnx.GetName())
	}
}
