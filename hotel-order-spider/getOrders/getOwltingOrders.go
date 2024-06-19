package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type GetOwltingOrderResponseBody struct {
	Data []struct {
		Order_serial string `json:"order_serial"`
		Order_status string `json:"order_status"`
		Created_at   string `json:"created_at"`
		Fullname     string `json:"fullname"`
		Room_names   string `json:"room_name"`
		Canceled_at  string `json:"canceled_at"`
		Sdate        string `json:"sdate"`
		Edate        string `json:"edate"`
		Source       string `json:"source"`
		Total        string `json:"total"`
	} `json:"data"`
	Pagination struct {
		Total_pages int `json:"total_pages"`
	} `json:"pagination"`
}

type GetOwltingOrderResponseBody2 struct {
	Data struct {
		Info struct {
			Order_serial     string `json:"order_serial"`
			Order_status     string `json:"order_status"`
			Orderer_fullname string `json:"orderer_fullname"`
			Source2          string `json:"order_source"`
			Source           string `json:"order_ota_full_name"`
			Sdate            string `json:"order_start_date"`
			Edate            string `json:"order_end_date"`
			Order_stay_night int    `json:"order_stay_night"`
		} `json:"info"`
		Rooms []struct {
			Room_name        string `json:"room_name"`
			Room_config_name string `json:"room_config_name"`
		} `json:"rooms"`
		Summary struct {
			Hotel struct {
				Receivable_total float64 `json:"receivable_total"`
				Paid_total       float64 `json:"paid_total"`
			} `json:"hotel"`
		} `json:"summary"`

		First_payment struct {
			Created_at string `json:"created_at"`
		} `json:"first_payment"`
	} `json:"data"`
}

type RoomInfo_owl struct {
	RoomType string
	Count    int
}

func GetOwlting(platform map[string]interface{}, dateFrom, dateTo, owltingAccommodationId string) {
	var result string
	var url string

	cookie, ok := platform["cookie"].(string)
	if !ok {
		fmt.Println("無法取得 cookie")
	}

	url = `https://www.owlting.com/booking/v2/admin/hotels/` + owltingAccommodationId + `/orders/calendar_list?lang=zh_TW&limit=1000&page=1&during_checkout_date=` + dateFrom + `,` + dateTo + `&order_by=id&sort_by=asc`

	fmt.Println("1.")
	if err := DoRequestAndGetResponse_owl("GET", url, http.NoBody, cookie, &result); err != nil {
		fmt.Println("DoRequestAndGetResponse failed!")
		fmt.Println("err", err)
		return
	}

	var ordersData GetOwltingOrderResponseBody
	err := json.Unmarshal([]byte(result), &ordersData)
	if err != nil {
		fmt.Println("JSON解碼錯誤:", err)
		return
	}
	pageCount := ordersData.Pagination.Total_pages
	fmt.Println("pageCount:", pageCount)

	var resultData []ReservationsDB
	var data ReservationsDB

	for _, reservation := range ordersData.Data {
		url = `https://www.owlting.com/booking/v2/admin/hotels/` + owltingAccommodationId + `/orders/` + reservation.Order_serial + `/detail?lang=zh_TW`

		if err := DoRequestAndGetResponse_owl("GET", url, http.NoBody, cookie, &result); err != nil {
			fmt.Println("DoRequestAndGetResponse failed!")
			fmt.Println("err", err)
			return
		}

		var orderData GetOwltingOrderResponseBody2
		err = json.Unmarshal([]byte(result), &orderData)
		if err != nil {
			fmt.Println("JSON解碼錯誤:", err)
			return
		}

		data.RoomNights = int64(orderData.Data.Info.Order_stay_night)

		roomInfoData := make(map[string]*RoomInfo_owl)
		for _, roomReservation := range orderData.Data.Rooms {
			roomType := roomReservation.Room_name

			roomInfo, ok := roomInfoData[roomType]
			if !ok {
				roomInfo = &RoomInfo_owl{
					RoomType: roomType,
					Count:    1,
				}
				roomInfoData[roomType] = roomInfo
			} else {
				roomInfo.Count++
			}
		}
		var combinedRoomInfo string
		for _, roomInfo := range roomInfoData {
			if combinedRoomInfo != "" {
				combinedRoomInfo += " + "
			}

			combinedRoomInfo += fmt.Sprintf("%s*%s", roomInfo.RoomType, strconv.Itoa(roomInfo.Count/int(orderData.Data.Info.Order_stay_night)))
		}

		data.BookingId = orderData.Data.Info.Order_serial

		data.GuestName = orderData.Data.Info.Orderer_fullname

		arrivalTime, err := time.Parse("2006-01-02", orderData.Data.Info.Sdate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}

		departureTime, err := time.Parse("2006-01-02", orderData.Data.Info.Edate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}

		parsedTime, err := time.Parse(time.RFC3339, orderData.Data.First_payment.Created_at)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return
		}

		resultTimeStr := parsedTime.Format("2006-01-02")
		data.BookDate = resultTimeStr

		checkOutTime := departureTime
		checkInTime := arrivalTime
		data.CheckOutDate = checkOutTime.Format("2006-01-02")
		data.CheckInDate = checkInTime.Format("2006-01-02")

		data.Price = float64(orderData.Data.Summary.Hotel.Receivable_total)
		if data.Price == 0 {
			data.Price = float64(orderData.Data.Summary.Hotel.Paid_total)
		}

		if orderData.Data.Info.Order_status == "fail_pay" || orderData.Data.Info.Order_status == "cancel" {
			data.ReservationStatus = "已取消"

		} else {
			data.ReservationStatus = "已成立"
		}

		if orderData.Data.Info.Source == "" {
			data.Platform = orderData.Data.Info.Source2

		} else {
			data.Platform = orderData.Data.Info.Source
		}

		data.Currency = "TWD"
		data.HotelId = owltingAccommodationId

		if data.Platform != "Booking.com" && data.Platform != "Agoda" && data.Platform != "CTrip" && data.Platform != "Expedia" && data.Platform != "Hostelworld" && data.Platform != "SiteMinder" && data.Platform != "manual" && data.Platform != "Airbnb" {
			resultData = append(resultData, data)
		}
	}
	fmt.Println("resultdata", resultData)

	fmt.Println(resultData)

	resultDataJSON, err := json.Marshal(resultData)
	if err != nil {
		fmt.Println("JSON 轉換錯誤:", err)
		return
	}

	var resultDB string
	// 將資料存入DB
	apiurl := "http://149.28.24.90:8893/revenue_reservation/setParseHtmlToDB"
	if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), cookie, &resultDB); err != nil {
		fmt.Println("setParseHtmlToDB failed!")
		return
	}

}

func DoRequestAndGetResponse_owl(method, postUrl string, reqBody io.Reader, cookie string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer cec478d3bca0a16b9b95b85f43096913cc9253ff80ac805dc546b9426f55e885")

	req.Header.Set("Cookie", cookie)
	switch resBody.(type) {
	case *string:
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

	// resBody of type *string is for html
	switch resBody := resBody.(type) {
	case *string:
		// If resBody is a string
		resBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*resBody = string(resBytes)
	default:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, resBody); err != nil {
			return err
		}
	}

	defer resp.Body.Close()

	return nil
}
