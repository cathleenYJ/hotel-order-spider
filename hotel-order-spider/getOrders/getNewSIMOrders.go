package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GetSiteOrderResponseBody struct {
	Data struct {
		Hotel struct {
			Spid                 string `json:"spid"`
			PlatformReservations struct {
				Results []struct {
					SourceId string `json:"sourceId"`
					Channel  struct {
						Name string `json:"name"`
					} `json:"channel"`
					FromDate     string `json:"fromDate"`
					CheckOutDate string `json:"checkOutDate"`
					Currency     string `json:"currency"`
					TotalAmount  struct {
						AmountAfterTax float64 `json:"amountAfterTax"`
					} `json:"totalAmount"`
					RoomStays []struct {
						RoomName string `json:"roomName"`
						Guests   []struct {
							FirstName string `json:"firstName"`
							LastName  string `json:"lastName"`
						} `json:"guests"`
					} `json:"roomStays"`
					Profiles []struct {
						FirstName string `json:"firstName"`
						LastName  string `json:"lastName"`
					} `json:"profiles"`
					PmsLastSentAt string `json:"pmsLastSentAt"`
					Type          string `json:"type"`
				} `json:"results"`
				Total int `json:"total"`
			} `json:"platformReservations"`
		} `json:"hotel"`
	} `json:"data"`
}

func GetNewSIM(platform map[string]interface{}, dateFrom, dateTo, newSIMAccommodationId string) {
	var result string

	cookie, ok := platform["cookie"].(string)
	if !ok {
		fmt.Println("cookie error")
	}
	x_xsrf_token, ok := platform["x_xsrf_token"].(string)
	if !ok {
		fmt.Println("x-xsrf-token error")
	}

	url := `https://platform.siteminder.com/api/cm-beef/graphql`

	var resultData []ReservationsDB
	requestJSON := `{"operationName":"getReservationsSearch","variables":{"spid":"` + newSIMAccommodationId + `","filters":{"checkOutDateRange":{"startDate":"` + dateFrom + `","endDate":"` + dateTo + `"}},"pagination":{"page":1,"pageSize":2000,"sortBy":"checkOutDate","sortOrder":"asc"}},"query":"query getReservationsSearch($spid: ID!, $filters: PlatformReservationsFilterInput, $pagination: PlatformReservationsPaginationInput) {\n  hotel(spid: $spid) {\n    spid\n    platformReservations(filters: $filters, pagination: $pagination) {\n      results {\n        uuid\n        sourceId\n        smPlatformReservationId\n        channel {\n          code\n          name\n          __typename\n        }\n        fromDate\n        toDate\n        checkOutDate\n        channelCreatedAt\n        currency\n        totalAmount {\n          amountAfterTax\n          amountBeforeTax\n          __typename\n        }\n        roomStays {\n          cmRoomRateUuid\n          cmChannelRoomRateUuid\n          roomName\n          numberOfAdults\n          numberOfChildren\n          numberOfInfants\n          guests {\n            companyName\n            firstName\n            middleName\n            lastName\n            __typename\n          }\n          __typename\n        }\n        guests {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        profiles {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        type\n        pmsDeliveryStatus\n        pmsLastSentAt\n        __typename\n      }\n      sortBy\n      sortOrder\n      page\n      pageSize\n      total\n      __typename\n    }\n    __typename\n  }\n}\n"}`

	if err := DoRequestAndGetResponse_sit("POST", url, strings.NewReader(requestJSON), cookie, x_xsrf_token, &result); err != nil {
		fmt.Println("DoRequestAndGetResponse failed!")
		fmt.Println("err", err)
		return
	}

	var ordersData GetSiteOrderResponseBody
	err := json.Unmarshal([]byte(result), &ordersData)
	if err != nil {
		fmt.Println("JSON解碼錯誤:", err)
		return
	}

	var data ReservationsDB
	for _, reservation := range ordersData.Data.Hotel.PlatformReservations.Results {
		data.BookingId = reservation.SourceId

		if len(reservation.Profiles) > 0 {
			data.GuestName = reservation.Profiles[0].FirstName + " " + reservation.Profiles[0].LastName
		} else if len(reservation.RoomStays) > 0 && len(reservation.RoomStays[0].Guests) > 0 {
			data.GuestName = reservation.RoomStays[0].Guests[0].FirstName + " " + reservation.RoomStays[0].Guests[0].LastName
		} else {
			data.GuestName = ""
		}

		arrivalTime, err := time.Parse("2006-01-02", reservation.FromDate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}

		departureTime, err := time.Parse("2006-01-02", reservation.CheckOutDate)
		if err != nil {
			fmt.Println("Error parsing departureTime time:", err)
		}

		var parsedTime time.Time
		if reservation.PmsLastSentAt == "" {
			parsedTime = time.Time{}
		} else {
			parsedTime, err = time.Parse(time.RFC3339, reservation.PmsLastSentAt)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				return
			}
		}

		roomStayCount := len(reservation.RoomStays)

		resultTimeStr := parsedTime.Format("2006-01-02")
		data.BookDate = resultTimeStr

		checkOutTime := departureTime
		checkInTime := arrivalTime
		data.CheckOutDate = checkOutTime.Format("2006-01-02")
		data.CheckInDate = checkInTime.Format("2006-01-02")

		duration := checkOutTime.Sub(checkInTime)
		roomNights := int64(duration.Hours() / 24)

		data.RoomNights = roomNights

		data.Price = reservation.TotalAmount.AmountAfterTax

		if reservation.Type == "Reservation" {
			data.ReservationStatus = "已成立"
		} else if reservation.Type == "Cancellation" {
			data.ReservationStatus = "已取消"
		}

		data.Platform = reservation.Channel.Name
		data.Currency = reservation.Currency
		data.HotelId = newSIMAccommodationId
		data.NumOfRooms = int64(roomStayCount)

		if data.Platform != "Booking.com" && data.Platform != "Agoda" && data.Platform != "Expedia" && data.Platform != "Trip.com(Old)" && data.Platform != "Trip.com (Old)" && data.Platform != "Hostelworld Group" {
			resultData = append(resultData, data)
		}

	}

	fmt.Println("resultData", resultData)

	resultDataJSON, err := json.Marshal(resultData)
	if err != nil {
		fmt.Println("JSON 轉換錯誤:", err)
		return
	}

	var resultDB string
	// 將資料存入DB
	apiurl := `http://149.28.24.90:8893/revenue_reservation/setParseHtmlToDB`
	if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), cookie, &resultDB); err != nil {
		fmt.Println("setParseHtmlToDB failed!")
		return
	}
	fmt.Println("resultDB:", resultDB)
}

func DoRequestAndGetResponse_sit(method, postUrl string, reqBody io.Reader, cookie string, x_xsrf_token string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("x-xsrf-token", x_xsrf_token)
	req.Header.Set("Cookie", cookie)

	fmt.Println("resBody", resBody)
	switch resBody := resBody.(type) {
	case *string:
		fmt.Println("string")
		fmt.Println("resBody", resBody)
		req.Header.Set("Content-Type", "application/json")
	default:
		fmt.Println("not string")
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 40 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println("resp", resp)

	// resBody of type *string is for html
	switch resBody := resBody.(type) {
	case *string:
		// If resBody is a string
		resBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*resBody = string(resBytes)

		fmt.Println("resBody", resBody)
	default:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, resBody); err != nil {
			return err
		}

		fmt.Println("data", data)
	}
	defer resp.Body.Close()
	return nil
}
