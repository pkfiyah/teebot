package teeonwrapper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkfiyah/tee1000/models"
	"github.com/redis/go-redis/v9"
)

type TeeOnClient struct {
	client *http.Client
	jar    *cookiejar.Jar
	redis  *redis.Client
	name   string
}

const teeOnSignInUrl string = "https://www.tee-on.com/PubGolf/servlet/com.teeon.teesheet.servlets.ajax.CheckSignInCloudAjax"
const teeOnTeeTimeUrl string = "https://www.tee-on.com/PubGolf/servlet/com.teeon.teesheet.servlets.golfersection.WebBookingBookTime"

const debug bool = true

func NewTeeOnClient() (*TeeOnClient, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	toClient := TeeOnClient{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	toClient.jar = jar
	toClient.client = &http.Client{
		Jar: jar,
	}
	toClient.redis = redisClient
	toClient.name = "test"

	return &toClient, nil
}

func (toc *TeeOnClient) TeeOnSignIn() error {

	form := url.Values{}
	form.Set("Username", "Tylerfancy")
	form.Set("Password", "MansoN666")
	form.Set("SaveSignIn", "false")
	form.Set("CourseCode", "")

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

func (toc *TeeOnClient) TeeOnSnipeTime(teeTime *models.TeeTime) error {
	// timestamping request
	nowTime := time.Now()
	unixT := nowTime.UnixMilli()

	for _, time := range teeTime.TimesToSnipe {
		formatTeeTime := time.Format("2006-01-02;15:04")
		fmt.Printf("Attempting time: %s\n", formatTeeTime)
		parts := strings.Split(formatTeeTime, ";")
		if len(parts) != 2 {
			return errors.New("Request time could not be parsed properly into a tee time")
		}

		form := url.Values{}

		form.Set(fmt.Sprintf("%d-0", unixT), "Tyler Fancy")
		for i := 1; i < int(teeTime.NumPlayers); i++ {
			form.Set(fmt.Sprintf("%d-%d", unixT, i), "Member")
		}
		form.Set("BackTarget", "com.teeon.teesheet.servlets.golfersection.WebBookingPlayerEntry")
		form.Set("CaptureCreditBluff", "false")
		form.Set("CaptureCreditMoneris", "false")
		form.Set("Carts", fmt.Sprintf("%d", teeTime.NumCarts))
		form.Set("CourseCode", "AVON")
		form.Set("Date", parts[0])
		form.Set("FromSpecials", "false")
		form.Set("Holes", fmt.Sprintf("%d", teeTime.NumHoles))
		form.Set("LockerString", "Tyler Fancy (PUB281288)1|0")
		form.Set("Name0", "Tyler Fancy")
		form.Set("PlayerID0", "AVON3971")
		for i := 1; i < int(teeTime.NumPlayers); i++ {
			form.Set(fmt.Sprintf("Name%d", i), "Member")
			form.Set(fmt.Sprintf("PlayerID%d", i), "")
		}
		form.Set("NineCode", "F")
		form.Set("Players", fmt.Sprintf("%d", teeTime.NumPlayers))
		form.Set("Referrer", "avonvalleygolf.com")
		form.Set("Ride0", "false")
		form.Set("Ride1", "false")
		form.Set("Ride2", "false")
		form.Set("Ride3", "false") // TODO Handle this based on 12 carts?
		form.Set("ShotgunID", "")
		form.Set("Time", parts[1])
		form.Set("UnlockTime", fmt.Sprintf("AVON|F|%s|%s|B|10:03|99", parts[0], parts[1]))

		req, err := http.NewRequest("POST", teeOnTeeTimeUrl, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := toc.client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp.StatusCode != 200 {
				err = fmt.Errorf("Non 200 response occured: %d:%s", resp.StatusCode, resp.Body)
			}
			return fmt.Errorf("Error occured while booking tee time or received non-200 response: %s", err)
		}

		defer resp.Body.Close()
		_, err = scanResponseForSnipeResult(resp.Body)

		if err != nil {
			fmt.Printf("Err Occ: %s\n", err)
			continue
		}
		fmt.Printf("No Error, Returning with time sniped")
		return nil
	}

	return nil
}

func scanResponseForSnipeResult(r io.Reader) (bool, error) {
	tooEarly, _ := regexp.Compile("You must wai[l-t]")
	notAvailable, _ := regexp.Compile("booking you requested is no longer available")
	maxBookings, _ := regexp.Compile("You have reached the maximum number of bookings")
	snipeSuccess, _ := regexp.Compile("Reservation Successful")

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if debug {
			fmt.Printf("[sRFSR]: %s\n", line)
		}

		if snipeSuccess.FindString(line) != "" {
			return false, nil
		}
		if tooEarly.FindString(line) != "" {
			return true, errors.New("Request too early, must wait for unlock")
		}
		if notAvailable.FindString(line) != "" {
			return true, errors.New("Booking not available")
		}
		if maxBookings.FindString(line) != "" {
			return true, errors.New("A tee time has already been booked for this date")
		}
	}

	return false, nil
}
