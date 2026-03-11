package brightness

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/misham/botpi/display"
)

type mockDisplay struct {
	brightness display.Brightness
}

func (m *mockDisplay) SetBrightness(b display.Brightness) {
	m.brightness = b
}

func TestControllerUpdate_NightIsDark(t *testing.T) {
	mock := &mockDisplay{}
	c := &Controller{
		display: mock,
		client:  &http.Client{},
		loc:     Location{Lat: 37.77, Lon: -122.43},
		now: func() time.Time {
			return time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC) // midnight PDT
		},
	}

	c.update()

	if mock.brightness != display.BrightnessDark {
		t.Errorf("got %d, want BrightnessDark", mock.brightness)
	}
}

func TestControllerUpdate_DaytimeClearIsBright(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"current":{"cloud_cover":20}}`)
	}))
	defer srv.Close()

	mock := &mockDisplay{}
	c := &Controller{
		display:    mock,
		client:     srv.Client(),
		loc:        Location{Lat: 37.77, Lon: -122.43},
		weatherURL: srv.URL,
		now: func() time.Time {
			return time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC) // noon PDT
		},
	}

	c.update()

	if mock.brightness != display.BrightnessBright {
		t.Errorf("got %d, want BrightnessBright", mock.brightness)
	}
}

func TestControllerUpdate_DaytimeOvercastDimsToNormal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"current":{"cloud_cover":85}}`)
	}))
	defer srv.Close()

	mock := &mockDisplay{}
	c := &Controller{
		display:    mock,
		client:     srv.Client(),
		loc:        Location{Lat: 37.77, Lon: -122.43},
		weatherURL: srv.URL,
		now: func() time.Time {
			return time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC) // noon PDT
		},
	}

	c.update()

	if mock.brightness != display.BrightnessNormal {
		t.Errorf("got %d, want BrightnessNormal", mock.brightness)
	}
}

func TestControllerUpdate_WeatherFailureFallsBackToSolar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	mock := &mockDisplay{}
	c := &Controller{
		display:    mock,
		client:     srv.Client(),
		loc:        Location{Lat: 37.77, Lon: -122.43},
		weatherURL: srv.URL,
		now: func() time.Time {
			return time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC) // noon PDT
		},
	}

	c.update()

	if mock.brightness != display.BrightnessBright {
		t.Errorf("got %d, want BrightnessBright (solar fallback)", mock.brightness)
	}
}

func TestControllerRun_StopsOnSignal(t *testing.T) {
	mock := &mockDisplay{}
	c := &Controller{
		display:  mock,
		client:   &http.Client{},
		loc:      Location{Lat: 37.77, Lon: -122.43},
		interval: time.Hour, // won't tick during test
		now: func() time.Time {
			return time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC)
		},
	}

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		c.Run(stop)
		close(done)
	}()

	close(stop)

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("Run did not stop within 1 second")
	}
}
