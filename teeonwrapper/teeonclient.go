package teeonwrapper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkfiyah/tee1000/models"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type TeeOnClient struct {
	client HTTPClient
}

const teeOnSignInUrl string = "https://www.tee-on.com/PubGolf/servlet/com.teeon.teesheet.servlets.ajax.CheckSignInCloudAjax"
const teeOnTeeTimeUrl string = "https://www.tee-on.com/PubGolf/servlet/com.teeon.teesheet.servlets.golfersection.WebBookingBookTime"

var ErrTooEarlyToRegisterTeeTime = errors.New("Request too early, must wait for unlock")
var ErrBookingNotAvailable = errors.New("Booking not available")
var ErrTeeTimeAlreadyBooked = errors.New("A tee time has already been booked for this date")

const debug bool = true

func NewTeeOnClient() (*TeeOnClient, error) {
	toClient := TeeOnClient{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	toClient.client = &http.Client{
		Jar: jar,
	}

	return &toClient, nil
}

func (toc *TeeOnClient) TeeOnSignIn(loginInfo *models.PlayerInfo) error {
	form := url.Values{}
	form.Set("Username", loginInfo.User.Username)
	form.Set("Password", loginInfo.User.Password)
	form.Set("CourseCode", loginInfo.Course.CourseCode)
	form.Set("SaveSignIn", "false")

	req, err := http.NewRequest("POST", teeOnSignInUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err = toc.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (toc *TeeOnClient) TeeOnSnipeTime(teeTime *models.TeeTime) (*time.Duration, error) {
	for _, snipeTime := range teeTime.TimesToSnipe {
		form, err := constructForm(teeTime, snipeTime)
		if err != nil {
			// Critical Parse Error
			return nil, err
		}

		req, err := http.NewRequest("POST", teeOnTeeTimeUrl, strings.NewReader(form.Encode()))
		if err != nil {
			// Critical Http Req Error
			return nil, err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := toc.client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp.StatusCode != 200 {
				err = fmt.Errorf("Non 200 response occured: %d:%s", resp.StatusCode, resp.Body)
			}
			return nil, fmt.Errorf("Error occured while booking tee time or received non-200 response: %s", err)
		}

		defer resp.Body.Close()
		retryIn, err := scanResponseForSnipeResult(resp.Body)

		// We got a retryIn time, booking will unlock soon
		if retryIn != nil {
			return retryIn, err
		}

		if err != nil {
			fmt.Printf("Err during result scan, continuing to next time?: %s\n", err)
			if err == ErrTooEarlyToRegisterTeeTime || err == ErrTeeTimeAlreadyBooked {
				return nil, err
			}

			// Booking not currently available, but can keep retrying to see if it frees up
			if err == ErrBookingNotAvailable {
				retryTime := time.Duration(float64(time.Second) * float64(math.Pow(2, float64(teeTime.Retries))))
				return &retryTime, err
			}
		}
	}

	return nil, nil
}

// Will scan POST results and determine booking success or failure.
// bool return describes whether we should retry booking at a later time.
func scanResponseForSnipeResult(r io.Reader) (*time.Duration, error) {
	tooEarly, _ := regexp.Compile(`You must wai[l-t]`)
	notAvailable, _ := regexp.Compile(`booking you requested is no longer available`)
	maxBookings, _ := regexp.Compile(`You have reached the maximum number of bookings`)
	snipeSuccess, _ := regexp.Compile(`Reservation Successful`)
	tooFarAhead, _ := regexp.Compile(`You cannot book this far ahead`)
	// `Monday, June 26, 2023 is the furthest day you are allowed to book.`
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if debug {
			fmt.Printf("[sRFSR]: %s\n", line)
		}

		// No need to retry, got booking successfully
		if snipeSuccess.FindString(line) != "" {
			return nil, nil
		}

		// Booking too far in the future to worry about retrying currently
		if tooFarAhead.FindString(line) != "" {
			return nil, ErrTooEarlyToRegisterTeeTime
		}

		// Booking will opening shortly, we will find out when and retry then.
		if tooEarly.FindString(line) != "" {
			// This response alludes to an unlock time within 24 hours. We will parse that time out and schedule booking for that time.
			unlockTime := regexp.MustCompile(`start booking for (?P<Datetime>.*am|.*pm)[.]`)
			res := unlockTime.FindStringSubmatch(line)
			if len(res) == 2 {
				const layout = "Monday, January 02, 2006 until 15:04 pm"
				parsedUnlockTime, terr := time.ParseInLocation(layout, res[1], time.Local)
				if terr != nil {
					return nil, terr
				}

				// date returned is for booking date, but the time is the unlock time. Assuming 7 days booking date difference
				updatedParsedUnlockTime := parsedUnlockTime.Add((-7 * (time.Hour * 24)))
				timeDiff := updatedParsedUnlockTime.Sub(time.Now())
				if timeDiff.Minutes() <= 5 {
					return &timeDiff, ErrTooEarlyToRegisterTeeTime
				}
			}

			return nil, ErrTooEarlyToRegisterTeeTime
		}

		// Booking not currently available, might be able to get it at a later time but well let retries handle that
		if notAvailable.FindString(line) != "" {
			return nil, ErrBookingNotAvailable
		}

		// We've alreayd booked a time that would conflict with this booking, nothing left to do here
		if maxBookings.FindString(line) != "" {
			return nil, ErrTeeTimeAlreadyBooked
		}
	}

	return nil, nil
}

func constructForm(teeTime *models.TeeTime, snipeTime time.Time) (*url.Values, error) {
	nowTime := time.Now()
	unixT := nowTime.UnixMilli()

	formatTeeTime := snipeTime.Format("2006-01-02;15:04")
	parts := strings.Split(formatTeeTime, ";")
	if len(parts) != 2 {
		return nil, errors.New("Request time could not be parsed properly into a tee time")
	}

	form := url.Values{}
	form.Set(fmt.Sprintf("%d-0", unixT), teeTime.BookingMember.User.Fullname)
	for i := 1; i < int(teeTime.NumPlayers); i++ {
		form.Set(fmt.Sprintf("%d-%d", unixT, i), "Member")
	}
	form.Set("BackTarget", "com.teeon.teesheet.servlets.golfersection.WebBookingPlayerEntry")
	form.Set("CaptureCreditBluff", "false")
	form.Set("CaptureCreditMoneris", "false")
	form.Set("Carts", fmt.Sprintf("%d", teeTime.NumCarts))
	form.Set("CourseCode", teeTime.BookingMember.Course.CourseCode)
	form.Set("Date", parts[0])
	form.Set("FromSpecials", "false")
	form.Set("Holes", fmt.Sprintf("%d", teeTime.NumHoles))
	form.Set("LockerString", teeTime.BookingMember.User.LockerString)
	form.Set("Name0", teeTime.BookingMember.User.Fullname)
	form.Set("PlayerID0", teeTime.BookingMember.User.PlayerID)
	for i := 1; i < int(teeTime.NumPlayers); i++ {
		form.Set(fmt.Sprintf("Name%d", i), "Member")
		form.Set(fmt.Sprintf("PlayerID%d", i), "")
	}
	form.Set("NineCode", "F")
	form.Set("Players", fmt.Sprintf("%d", teeTime.NumPlayers))
	form.Set("Referrer", teeTime.BookingMember.Course.Referrer)

	form.Set("Ride0", "false")
	form.Set("Ride1", "false")
	form.Set("Ride2", "false")
	form.Set("Ride3", "false") // TODO Handle this based on 12 carts?
	form.Set("ShotgunID", "")
	form.Set("Time", parts[1])
	form.Set("UnlockTime", fmt.Sprintf("%s|F|%s|%s|B|10:03|99", teeTime.BookingMember.Course.CourseCode, parts[0], parts[1]))

	return &form, nil
}
