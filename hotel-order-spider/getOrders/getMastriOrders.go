package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type GetMastriOrderResponseBody []struct {
	Source      string `json:"source"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Number      string `json:"number"`
	Name        string `json:"name"`
	CheckIn     string `json:"check_in"`
	CheckOut    string `json:"check_out"`
	BookingRoom string `json:"booking_room"`
	CheckInRoom string `json:"check_in_room"`
	Rooms       int    `json:"rooms"`
	Price       string `json:"price"`
	CreatedAt   string `json:"created_at"`
	CanceledAt  string `json:"canceled_at"`
}

func GetMastri(platform map[string]interface{}, dateFrom, dateTo string) {
	var result string

	hotels, ok := platform["hotel"].([]interface{})
	if !ok || hotels == nil {
		fmt.Println("hotel error")
	}

	for _, hotelRaw := range hotels {
		hotel, ok := hotelRaw.(map[string]interface{})
		if !ok || hotel == nil {
			fmt.Println("hotel error")
			continue
		}

		hotelName, ok := hotel["name"].(string)
		if !ok {
			fmt.Println("hotel name error")
			continue
		}
		fmt.Printf("hotelName: %s", hotelName)

		url := "http://mrhost.xcodemy.com/api/vendor/getMasterOrders"

		var resultData []ReservationsDB
		requestJSON := `{"domain": "` + hotelName + `" ,"date_type": "check_out","start_date": "` + dateFrom + `" ,"end_date": "` + dateTo + `"}`
		jsonReqBody := []byte(requestJSON)
		if err := DoRequestAndGetResponse("POST", url, bytes.NewBuffer(jsonReqBody), "", &result); err != nil {
			fmt.Println("DoRequestAndGetResponse failed!")
			fmt.Println("err", err)
			return
		}

		var ordersData GetMastriOrderResponseBody
		err := json.Unmarshal([]byte(result), &ordersData)
		if err != nil {
			fmt.Println("JSON解碼錯誤:", err)
			return
		}

		fmt.Println("ordersData", ordersData)

		var data ReservationsDB
		for _, reservation := range ordersData {
			data.Platform = reservation.Source
			data.GuestName = reservation.Name
			data.BookDate = reservation.CreatedAt
			data.CheckOutDate = reservation.CheckOut
			data.CheckInDate = reservation.CheckIn

			if reservation.Status == "CANCELED" {
				data.ReservationStatus = "已取消"
			} else if reservation.Status == "CHECKED_OUT" || reservation.Status == "NO_SHOW" {
				data.ReservationStatus = "已成立"
			}
			// UPCOMING 即將入住
			// CANCELED 已取消
			// CHECKED_IN 已入住
			// CHECKED_OUT 已退房
			// NO_SHOW

			data.BookingId = reservation.Number
			data.Currency = "TWD"

			price, _ := strconv.ParseFloat(reservation.Price, 64)
			data.Price = price

			startDate, _ := time.Parse("2006-01-02", reservation.CheckIn)
			endDate, _ := time.Parse("2006-01-02", reservation.CheckOut)
			roomNights := 0
			for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
				roomNights += 1
			}
			data.RoomNights = int64(roomNights) - 1

			if hotelName == "guidetpedadaocheng" {
				data.HotelId = "R10052"
			} else if hotelName == "getchahostel" {
				data.HotelId = "R10059"
			} else if hotelName == "hamiltonhotel" {
				data.HotelId = "R10044"
			} else if hotelName == "dreammansionhotel" {
				data.HotelId = "R10048"
			}

			if data.Platform == "線上" {
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
		if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), "", &resultDB); err != nil {
			fmt.Println("setParseHtmlToDB failed!")
			return
		}
		fmt.Println("resultDB:", resultDB)
	}
}
