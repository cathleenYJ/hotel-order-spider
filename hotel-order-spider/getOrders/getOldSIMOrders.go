package getOrders

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GetOldSIMOrderResponseBody struct {
	TotalReservations int `json:"totalReservations"`
	Max               int `json:"max"`
	Offset            int `json:"offset"`
	Lower             int `json:"lower"`
	Upper             int `json:"upper"`
	Reservations      []struct {
		ID                    int    `json:"id"`
		PlatformReservationID string `json:"platformReservationId"`
		UUID                  string `json:"uuid"`
		HotelierID            int    `json:"hotelierId"`
		SourceID              string `json:"sourceId"`
		SiteminderID          string `json:"siteminderId"`
		Status                string `json:"status"`
		ChannelName           string `json:"channelName"`
		CreatedAt             string `json:"createdAt"`
		CheckIn               string `json:"checkIn"`
		CheckOut              string `json:"checkOut"`
		Guest                 struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"guest"`
		Total                  float64 `json:"total"`
		Currency               string  `json:"currency"`
		HotelierTimeZoneOffset string  `json:"hotelierTimeZoneOffset"`
	} `json:"reservations"`
}

func GetOldSIM(platform map[string]interface{}, dateFrom, dateTo, oldSIMAccommodationId string) {
	var result string
	var resultData []ReservationsDB
	var ordersData GetOldSIMOrderResponseBody

	cookie, _ := platform["cookie"].(string)
	x_xsrf_token, _ := platform["x_xsrf_token"].(string)
	x_xsrf_token_url, _ := platform["x_xsrf_token_url"].(string)

	url := fmt.Sprintf("https://app-apac.siteminder.com/web/extranet/reloaded/hoteliers/%s/reservations", oldSIMAccommodationId)
	reqBodyStr := fmt.Sprintf("{\"dateType\":\"checkout\",\"orderBy\":{\"columnName\":\"dateCreated\",\"order\":\"asc\"},\"reservationStatus\":{},\"channels\":{},\"fromDate\":\"%s\",\"toDate\":\"%s\",\"offset\":0}", dateFrom, dateTo)
	jsonReqBody := []byte(reqBodyStr)
	if err := DoRequestAndGetResponse_oldSIM("POST", url, bytes.NewBuffer(jsonReqBody), cookie, x_xsrf_token, x_xsrf_token_url, &ordersData); err != nil {
		fmt.Println("DoRequestAndGetResponse failed!")
		fmt.Println("err", err)
		return
	}

	total := ordersData.TotalReservations
	offset := ordersData.Offset

	fmt.Println("total,offset", total, offset)
	for offset < total {
		fmt.Println(" inner total,offset", total, offset)
		// Send a request.
		reqBodyStr := fmt.Sprintf("{\"dateType\":\"checkout\",\"orderBy\":{\"columnName\":\"dateCreated\",\"order\":\"asc\"},\"reservationStatus\":{},\"channels\":{},\"fromDate\":\"%s\",\"toDate\":\"%s\",\"offset\":%d}", dateFrom, dateTo, offset)
		jsonReqBody := []byte(reqBodyStr)
		if err := DoRequestAndGetResponse_oldSIM("POST", url, bytes.NewBuffer(jsonReqBody), cookie, x_xsrf_token, x_xsrf_token_url, &ordersData); err != nil {
			fmt.Println("DoRequestAndGetResponse failed!")
			fmt.Println("err", err)
			return
		}

		fmt.Println("result", result)

		for _, reservation := range ordersData.Reservations {
			var data ReservationsDB
			data.Platform = reservation.ChannelName
			data.GuestName = reservation.Guest.FirstName + " " + reservation.Guest.LastName

			// 解析時間字串為時間格式
			t, err := time.Parse(time.RFC3339, reservation.CreatedAt)
			if err != nil {
				fmt.Println("Parse error:", err)
			}
			// 格式化時間為日期格式（YYYY-MM-DD）
			data.BookDate = t.Format("2006-01-02")

			data.CheckInDate = reservation.CheckIn
			data.CheckOutDate = reservation.CheckOut

			originalStatus := strings.TrimPrefix(reservation.Status, "app.reservations.status.")
			if originalStatus == "booked" || originalStatus == "modified" {
				data.ReservationStatus = "已成立"
			} else if originalStatus == "cancelled" {
				data.ReservationStatus = "已取消"
			} else {
				data.ReservationStatus = "null"
			}

			data.BookingId = reservation.SourceID
			data.Currency = reservation.Currency
			data.Price = reservation.Total

			startDate, _ := time.Parse("2006-01-02", reservation.CheckIn)
			endDate, _ := time.Parse("2006-01-02", reservation.CheckOut)
			roomNights := 0
			for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
				roomNights += 1
			}
			data.RoomNights = int64(roomNights) - 1

			data.HotelId = oldSIMAccommodationId

			if reservation.CheckIn == "" || reservation.CheckOut == "" {
				fmt.Println("data.BookingId", data.BookingId)
			}

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
		offset += 15
	}
}

func DoRequestAndGetResponse_oldSIM(method string, url string, reqBody io.Reader, cookie string, token string, tokenUri string, resBody interface{}) error {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", cookie)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-XSRF-TOKEN", token)
	req.Header.Set("X-XSRF-TOKEN-URI", tokenUri)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Read response and trim the garbage prefixes.
	data, _ := io.ReadAll(resp.Body)
	dataString := string(data)

	trimmedData := strings.TrimPrefix(dataString, `)]}',`)

	// If set price succeeded, response should be nil.
	if method == "PUT" && (dataString != "" && dataString != "\n") {
		fmt.Println("Set price returned unexpected response!")
		return errors.New("set price request failed")
	}

	if resBody != nil {
		if err := json.Unmarshal([]byte(trimmedData), resBody); err != nil {
			fmt.Println("123Unmarshal error!")
			return err
		}
	}

	defer resp.Body.Close()

	return nil
}
