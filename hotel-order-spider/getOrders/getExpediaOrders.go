package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetExpedia(platform map[string]interface{}, dateFrom, dateTo string) {

	cookie, ok := platform["cookie"].(string)
	if !ok {
		fmt.Println("cookie error")
	}

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

		hotelId, ok := hotel["hotelid"].(string)
		if !ok {
			fmt.Println("hotel id error")
			continue
		}

		url := fmt.Sprintf("https://ycs.agoda.com/zh-tw/%s/kipp/api/hotelBookingApi/Get", hotelId)

		var resultData []ReservationsDB
		var ordersData GetAgodaOrderResponseBody

		startDateTime, _ := time.Parse("2006-01-02", dateFrom)
		endDateTime, _ := time.Parse("2006-01-02", dateTo)

		startDateUnixTime := startDateTime.UnixMilli()
		endDateUnixTime := endDateTime.UnixMilli()

		if err := PostForAgodaReservations(url, startDateUnixTime, endDateUnixTime, cookie, &ordersData); err != nil {
			return
		}

		wg := new(sync.WaitGroup)

		for _, reservation := range ordersData.Bookings {
			checkOutDate, _ := time.Parse("2006-01-02", strings.Split(reservation.CheckOutDate, "T")[0])

			// check_out_date out of range.
			if checkOutDate.Before(startDateTime) || checkOutDate.After(endDateTime) {
				continue
			}

			// If valid, do a API call.
			postUrl := fmt.Sprintf("https://ycs.agoda.com/en-us/%s/kipp/api/hotelBookingApi/GetDetails", hotelId)

			wg.Add(1)

			var resultDetail GetAgodaOrderDetailsResponseBody
			if err := PostForAgodaReservationsDetails(postUrl, reservation.BookingID, cookie, &resultDetail, wg); err != nil {
				fmt.Println("PostForReservationsDetails failed!")
				return
			}

			wg.Add(1)

			defer timeTrack(time.Now(), "ArrangeReservationData")
			defer wg.Done()

			var arrangedData ReservationsDB

			arrangedData.Platform = "Agoda"
			arrangedData.BookingId = strconv.Itoa(int(reservation.BookingID))
			arrangedData.GuestName = reservation.GuestName
			arrangedData.NumOfGuests = reservation.Adults + reservation.Children
			arrangedData.HotelId = hotelId

			for _, bookingDetail := range resultDetail.Data {

				bookingDateStr, _ := time.Parse("2006-01-02", strings.Split(bookingDetail.APIData.BookingDate, "T")[0])
				arrangedData.BookDate = bookingDateStr.Format("2006-01-02")
				checkInDateStr, _ := time.Parse("2006-01-02", strings.Split(bookingDetail.APIData.CheckInDate, "T")[0])
				arrangedData.CheckInDate = checkInDateStr.Format("2006-01-02")
				checkOutDateStr, _ := time.Parse("2006-01-02", strings.Split(bookingDetail.APIData.CheckOutDate, "T")[0])
				arrangedData.CheckOutDate = checkOutDateStr.Format("2006-01-02")

				arrangedData.NumOfRooms = bookingDetail.APIData.NoOfRoom
				arrangedData.Price = bookingDetail.APIData.RateDetailList.TotalNetInclusive
				arrangedData.Currency = bookingDetail.APIData.RateDetailList.Currency

				originalStatus := bookingDetail.APIData.AckRequestType
				fmt.Println("originalStatus", originalStatus)
				if originalStatus == 1 || originalStatus == 3 {
					arrangedData.ReservationStatus = "已成立"
				} else if originalStatus == 2 {
					arrangedData.ReservationStatus = "已取消"
					if bookingDetail.APIData.RateDetailList.TotalNetInclusive != 0 {
						arrangedData.ReservationStatus = "Chargeable cancellation"
					}
				} else {
					stringValue := strconv.FormatInt(bookingDetail.APIData.AckRequestType, 10)
					arrangedData.ReservationStatus = stringValue
				}

				start, _ := time.Parse("2006-01-02", arrangedData.CheckInDate)
				end, _ := time.Parse("2006-01-02", arrangedData.CheckOutDate)

				roomNights := 0
				for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
					roomNights += 1
				}
				arrangedData.RoomNights = int64(roomNights) - 1

				messages := ""

				for _, message := range bookingDetail.APIData.MessageList {
					messages = messages + message.MessageProperty + "\n"
				}
				arrangedData.GuestRequest = messages
			}
			resultData = append(resultData, arrangedData)
		}

		fmt.Println("resultData", resultData)

		resultDataJSON, err := json.Marshal(resultData)
		if err != nil {
			fmt.Println("JSON 轉換錯誤:", err)
			return
		}

		if len(resultData) > 0 {
			var resultDB string
			// 將資料存入DB
			apiurl := `http://149.28.24.90:8893/revenue_reservation/setParseHtmlToDB`
			if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), cookie, &resultDB); err != nil {
				fmt.Println("setParseHtmlToDB failed!")
				return
			}
		}
	}
}
