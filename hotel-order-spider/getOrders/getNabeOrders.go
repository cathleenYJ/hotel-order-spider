package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	// "strconv"
	"time"
)

type GetNabeReservationsResponseBody struct {
	ResNo   int    `json:"ResNo"`
	Status  string `json:"Status"`
	Content struct {
		Orders map[string]struct {
			Orders []struct {
				ResEleID      string `json:"ResEleID"`
				ResID         string `json:"ResID"`
				Status        string `json:"Status"`
				CustFName     string `json:"CustFName"`
				CustLName     string `json:"CustLName"`
				Country       string `json:"Country"`
				GuestFName    string `json:"GuestFName"`
				GuestLName    string `json:"GuestLName"`
				OtaBookID     string `json:"OtaBookID"`
				ResUpdateTime string `json:"ResUpdateTime"`
				OtaEleID      string `json:"OtaEleID"`
				RoomID        string `json:"RoomID"`
				OtaID         string `json:"OtaID"`
				Arrival       string `json:"Arrival"`
				Depart        string `json:"Depart"`
				Charge        string `json:"Charge"`
				CreateTime    string `json:"CreateTime"`
				UpdateTime    string `json:"UpdateTime"`
				SpecialRQ     string `json:"SpecialRQ"`
				People        string `json:"People"`
				Qty           string `json:"Qty"`
				Handle        string `json:"Handle"`
				OtaName       string `json:"OtaName"`
				RoomName      string `json:"RoomName"`
				PMSResult     string `json:"PMS_Result"`
				Super         bool   `json:"super"`
			} `json:"orders"`
		} `json:"orders"`
	} `json:"Content"`
}

func GetNabe(platform map[string]interface{}, dateFrom, dateTo, nabeAccommodationId, hotelName, mrhostId string) {

	fmt.Println()
	fmt.Println(hotelName, mrhostId, nabeAccommodationId)

	var resultData []ReservationsDB

	cookie, _ := platform["cookie"].(string)
	url := "https://www.hotelnabe.com.tw/order/search_order?date_start=" + dateFrom + "&date_end=" + dateTo + "&search_type=depart&ota=0&customer_name=&order_no=&res_id=&_=1723515339163"

	var ordersData GetNabeReservationsResponseBody
	if err := DoRequestAndGetResponse_nabe("POST", url, http.NoBody, cookie, &ordersData); err != nil {
		fmt.Println("err:", err)
		return
	}

	if ordersData.Status != "Success" {
		fmt.Println()
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("! 請更新config_nabe.yaml中的cookie !")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println()
		os.Exit(1)
	}

	for _, dateOrders := range ordersData.Content.Orders {
		for _, order := range dateOrders.Orders {
			var data ReservationsDB
			data.BookingId = order.OtaBookID
			data.GuestName = order.CustFName + " " + order.CustLName
			data.Platform = "Nabe"
			data.CheckInDate = order.Arrival
			data.CheckOutDate = order.Depart

			bookDate, err := time.Parse("2006-01-02 15:04:05", order.ResUpdateTime)
			if err != nil {
				fmt.Println("Error parsing book date:", err)
				return
			}
			data.BookDate = bookDate.Format("2006-01-02")

			data.RoomType = order.RoomName
			price, err := strconv.ParseFloat(order.Charge, 64)
			if err != nil {
				fmt.Println("Error converting price:", err)
				return
			}
			data.Price = price
			data.Commission = 0
			data.Currency = "TWD"

			originalStatus := order.Status
			if originalStatus == "一般" || originalStatus == "修改" {
				data.ReservationStatus = "已成立"
			} else if originalStatus == "取消" {
				data.ReservationStatus = "已取消"
				if data.Price != 0 {
					data.ReservationStatus = "Chargeable cancellation"
				}
			} else {
				data.ReservationStatus = originalStatus
			}

			startDate, _ := time.Parse("2006-01-02", order.Arrival)
			endDate, _ := time.Parse("2006-01-02", order.Depart)
			roomNights := 0
			for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
				roomNights += 1
			}
			data.RoomNights = int64(roomNights) - 1
			numOfGuests, err := strconv.ParseInt(order.People, 10, 64)
			if err != nil {
				fmt.Println("Error converting numOfGuests:", err)
				return
			}
			data.NumOfGuests = numOfGuests

			data.HotelId = nabeAccommodationId

			if order.OtaName != "Agoda" && order.OtaName != "Booking" && order.OtaName != "Expedia" && order.OtaName != "Ctrip-CM預付" {
				resultData = append(resultData, data)
			}
		}
	}
	fmt.Println("len(resultData)", len(resultData))
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
			fmt.Println("nabe setParseHtmlToDB failed!")
			return
		}
	}
}

func DoRequestAndGetResponse_nabe(method string, url string, reqBody io.Reader, cookie string, resBody interface{}) error {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", " ")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Name", "pc-reservations-web")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, resBody); err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
