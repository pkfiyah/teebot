package teeonwrapper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pkfiyah/tee1000/models"
)

type TestClient struct {
	clientMock     HTTPClient
	expectError    bool
	expectDuration bool
}

var testingClients = []TestClient{
	{
		clientMock:     HttpClientMockSuccess{},
		expectError:    false,
		expectDuration: false,
	},
	{
		clientMock:     HttpClientMockTooFarAhead{},
		expectError:    true,
		expectDuration: false,
	},
	{
		clientMock:     HttpClientMockNotAvailable{},
		expectError:    true,
		expectDuration: false,
	},
	{
		clientMock:     HttpClientMockMaxBookings{},
		expectError:    true,
		expectDuration: false,
	},
	{
		clientMock:     HttpClientMockTooEarly{},
		expectError:    true,
		expectDuration: false,
	},
	{
		clientMock:     HttpClientMockNearlyTime{},
		expectError:    true,
		expectDuration: true,
	},
}

type HttpClientMockSuccess struct{}

func (c HttpClientMockSuccess) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Status:     "200",
		Body:       ioutil.NopCloser(strings.NewReader("Reservation Successful")),
	}, nil
}

type HttpClientMockTooFarAhead struct{}

func (c HttpClientMockTooFarAhead) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Status:     "200",
		Body:       ioutil.NopCloser(strings.NewReader("You cannot book this far ahead")),
	}, nil
}

type HttpClientMockNotAvailable struct{}

func (c HttpClientMockNotAvailable) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Status:     "200",
		Body:       ioutil.NopCloser(strings.NewReader("booking you requested is no longer available")),
	}, nil
}

type HttpClientMockMaxBookings struct{}

func (c HttpClientMockMaxBookings) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Status:     "200",
		Body:       ioutil.NopCloser(strings.NewReader("You have reached the maximum number of bookings")),
	}, nil
}

type HttpClientMockNearlyTime struct{}

func (c HttpClientMockNearlyTime) Do(req *http.Request) (*http.Response, error) {
	// Getting time in a format return from Tee On from Avon valley, unlock time of 7 days prior at 6pm
	// Will return current time + 7 days and 2 hours
	formatTime := time.Now().Add((time.Hour * 24 * 7) + time.Minute*2).Format("Monday, January 02, 2006 until 15:04 pm")

	return &http.Response{
		StatusCode: 200,
		Status:     "200",
		Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf("You must wait |||| start booking for %s.", formatTime))),
	}, nil
}

type HttpClientMockTooEarly struct{}

func (c HttpClientMockTooEarly) Do(req *http.Request) (*http.Response, error) {
	// Getting time in a format return from Tee On from Avon valley, unlock time of 7 days prior at 6pm
	// Will return current time + 7 days and 2 hours
	formatTime := time.Now().Add((time.Hour * 24 * 7) + time.Hour*2).Format("Monday, January 2, 2006 until 15:04 pm")

	return &http.Response{
		StatusCode: 200,
		Status:     "200",
		Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf("You must wait |||| start booking for %s.", formatTime))),
	}, nil
}

// Runs through a set of test clients with mock returns to match regex
func Test_TestClientArray(t *testing.T) {
	for _, client := range testingClients {
		toClient, err := NewTeeOnClient()
		toClient.client = client.clientMock

		duration, err := toClient.TeeOnSnipeTime(getMockTeeTime())
		if client.expectError && err == nil {
			t.Error("Expected error, got none:")
			return
		}

		if client.expectDuration && duration == nil {
			t.Error("Expected Duration, got none:")
			return
		}

		if err != nil {
			t.Log(err)
			return
		}
	}
}

func Test_TeeOnSnipeTimeSuccess(t *testing.T) {
	toClient, err := NewTeeOnClient()
	toClient.client = &HttpClientMockSuccess{}
	if err != nil {
		t.Errorf("Couldn't create Tee On client")
	}

	_, err = toClient.TeeOnSnipeTime(getMockTeeTime())
	if err != nil {
		t.Error(err)
		return
	}
}

func getMockTeeTime() *models.TeeTime {
	return &models.TeeTime{
		BookingMember: getMockPlayerInfo(),
		BookingDate:   "2001-09-11",
		TimesToSnipe:  []time.Time{time.Date(2001, 9, 11, 8, 46, 0, 0, time.Local), time.Date(2001, 9, 11, 9, 3, 0, 0, time.Local)},
		NumPlayers:    4,
		NumCarts:      0,
		NumHoles:      18,
		Retries:       10,
		RedKey:        "TeeTime:*some other random stuff here",
	}
}

func getMockPlayerInfo() *models.PlayerInfo {
	return &models.PlayerInfo{
		User: models.UserInfo{
			Username:     "RyuHadoken",
			Fullname:     "Ryu Hadoken",
			Password:     "Sureyoucan",
			PlayerID:     "RYU1094",
			LockerString: "Ryu Hadoken 12123(3432) 1|0",
		},
		Course: models.CourseInfo{
			CourseCode: "AABBUDUD",
			Referrer:   "hadokenadacademy.com",
		},
	}
}
