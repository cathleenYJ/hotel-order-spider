package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GetAgodaOrderResponseBody struct {
	Bookings []struct {
		BookingDate      string `json:"BookingDate"`
		BookingID        int64  `json:"BookingId"`
		CheckInDate      string `json:"CheckInDate"`
		CheckOutDate     string `json:"CheckOutDate"`
		GuestName        string `json:"GuestName"`
		PaymentModel     int64  `json:"PaymentModel"`
		RatePlan         int64  `json:"RatePlan"`
		StatusChangeDate string `json:"StatusChangeDate"`
		EmailType        int64  `json:"EmailType"`
		Adults           int64  `json:"Adults"`
		Children         int64  `json:"Children"`
		RoomTypeID       int64  `json:"RoomTypeId"`
	} `json:"Bookings"`
}

type GetAgodaOrderDetailsResponseBody struct {
	Data []struct {
		FirstName  string `json:"FirstName"`
		LastName   string `json:"LastName"`
		RoomNights int64  `json:"RoomNights"`
		APIData    struct {
			AckRequestType   int64  `json:"AckRequestType"`
			Adult            int64  `json:"Adult"`
			BookingDate      string `json:"BookingDate"`
			BookingDateStr   string `json:"BookingDateStr"`
			BookingTimeStr   string `json:"BookingTimeStr"`
			BookingID        int64  `json:"BookingId"`
			Channel          int64  `json:"Channel"`
			ChannelName      string `json:"ChannelName"`
			CheckInDate      string `json:"CheckInDate"`
			CheckInDateStr   string `json:"CheckInDateStr"`
			CheckOutDate     string `json:"CheckOutDate"`
			CheckOutDateStr  string `json:"CheckOutDateStr"`
			Child            int64  `json:"Child"`
			GuestName        string `json:"GuestName"`
			GuestNationality string `json:"GuestNationality"`
			NoOfRoom         int64  `json:"NoOfRoom"`
			Occupancy        int64  `json:"Occupancy"`
			RateDetailList   struct {
				DailyRates []struct {
					Date                    string  `json:"Date"`
					NetExclusive            any     `json:"NetExclusive"`
					NetFee                  any     `json:"NetFee"`
					NetInclusive            float64 `json:"NetInclusive"`
					NetTax                  any     `json:"NetTax"`
					SellExclusive           any     `json:"SellExclusive"`
					SellFee                 any     `json:"SellFee"`
					SellInclusive           any     `json:"SellInclusive"`
					SellTax                 any     `json:"SellTax"`
					SurchargeID             int     `json:"SurchargeId"`
					Type                    int     `json:"Type"`
					TypeName                string  `json:"TypeName"`
					ReferenceSellInclusive  any     `json:"ReferenceSellInclusive"`
					Commission              any     `json:"Commission"`
					WithholdingTax          any     `json:"WithholdingTax"`
					TaxOnCommission         any     `json:"TaxOnCommission"`
					NormalAndWithholdingTax any     `json:"NormalAndWithholdingTax"`
					ApmApprovalPriceID      any     `json:"ApmApprovalPriceId"`
					AiAdjusted              any     `json:"AiAdjusted"`
					DateString              string  `json:"DateString"`
					DateStr                 string  `json:"DateStr"`
				} `json:"DailyRates"`
				Currency                    string  `json:"Currency"`
				TotalNetExclusive           any     `json:"TotalNetExclusive"`
				TotalNetFee                 any     `json:"TotalNetFee"`
				TotalNetInclusive           float64 `json:"TotalNetInclusive"`
				TotalNetTax                 any     `json:"TotalNetTax"`
				TotalReferenceSellInclusive any     `json:"TotalReferenceSellInclusive"`
				TotalSellExclusive          any     `json:"TotalSellExclusive"`
				TotalSellFee                any     `json:"TotalSellFee"`
				TotalSellInclusive          any     `json:"TotalSellInclusive"`
				TotalSellTax                any     `json:"TotalSellTax"`
			} `json:"RateDetailList"`
			MessageList []struct {
				ID             string `json:"Id"`
				CreatedAt      string `json:"CreatedAt"`
				ConversationID string `json:"ConversationId"`
				Guest          struct {
					ID         int    `json:"Id"`
					Email      string `json:"Email"`
					FirstName  string `json:"FirstName"`
					LastName   string `json:"LastName"`
					Language   string `json:"Language"`
					LanguageID int    `json:"LanguageId"`
				} `json:"Guest"`
				Sender     string `json:"Sender"`
				Target     string `json:"Target"`
				Properties struct {
					MessageType string `json:"MessageType"`
				} `json:"Properties"`
				MessageProperty string `json:"MessageProperty"`
				SpecialRequests []struct {
					CmsID int    `json:"CmsId"`
					Name  string `json:"Name"`
					Value any    `json:"Value"`
				} `json:"SpecialRequests"`
				Status     int `json:"Status"`
				StatusInfo any `json:"StatusInfo"`
			} `json:"MessageList"`
			RateCategoryName       string `json:"RateCategoryName"`
			RoomtypeID             int64  `json:"RoomtypeId"`
			RoomtypeAlternateName  string `json:"RoomtypeAlternateName"`
			RoomtypeName           string `json:"RoomtypeName"`
			SpecialRequest         string `json:"SpecialRequest"`
			StatusChangeDate       string `json:"StatusChangeDate"`
			StatusChangeDateStr    string `json:"StatusChangeDateStr"`
			IsConfirmRequest       bool   `json:"IsConfirmRequest"`
			BookingDateString      string `json:"BookingDateString"`
			CheckInDateString      string `json:"CheckInDateString"`
			CheckOutDateString     string `json:"CheckOutDateString"`
			StatusChangeDateString string `json:"StatusChangeDateString"`
		} `json:"ApiData"`
	} `json:"Data"`
}

func GetAgoda(platform map[string]interface{}, dateFrom, dateTo, agodaAccommodationId, hotelName, mrhostId string) {

	fmt.Println()
	fmt.Println(hotelName, mrhostId, agodaAccommodationId)
	fmt.Println()

	var resultData []ReservationsDB
	var ordersData GetAgodaOrderResponseBody

	cookie, _ := platform["cookie"].(string)
	startDateTime, _ := time.Parse("2006-01-02", dateFrom)
	endDateTime, _ := time.Parse("2006-01-02", dateTo)
	startDateUnixTime := startDateTime.UnixMilli()
	endDateUnixTime := endDateTime.UnixMilli()

	url := fmt.Sprintf("https://ycs.agoda.com/zh-tw/%s/kipp/api/hotelBookingApi/Get", agodaAccommodationId)
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
		postUrl := fmt.Sprintf("https://ycs.agoda.com/en-us/%s/kipp/api/hotelBookingApi/GetDetails", agodaAccommodationId)

		wg.Add(1)

		var resultDetail GetAgodaOrderDetailsResponseBody
		if err := PostForAgodaReservationsDetails(postUrl, reservation.BookingID, cookie, &resultDetail, wg); err != nil {
			fmt.Println()
			fmt.Println("!!!!!!!!!!!!!!!")
			fmt.Println("! 請更新 cookie!")
			fmt.Println("!!!!!!!!!!!!!!!")
			return
		}

		fmt.Println("reservation.BookingID", reservation.BookingID)
		wg.Add(1)
		defer wg.Done()

		var arrangedData ReservationsDB

		arrangedData.Platform = "Agoda"
		arrangedData.BookingId = strconv.Itoa(int(reservation.BookingID))
		arrangedData.GuestName = reservation.GuestName
		arrangedData.NumOfGuests = reservation.Adults + reservation.Children
		arrangedData.HotelId = agodaAccommodationId

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

			time.Sleep(1 * time.Second)
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

func PostForAgodaReservations(postUrl string, startDate int64, endDate int64, cookie string, resBody interface{}) error {

	urlData := url.Values{}
	urlData.Set("UseCheckinDate", "true")
	urlData.Set("CheckInDateFromJson", fmt.Sprintf("/Date(%d)/", startDate))
	urlData.Set("CheckInDateToJson", fmt.Sprintf("/Date(%d)/", endDate))

	req, err := http.NewRequest("POST", postUrl, strings.NewReader(urlData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 60 * time.Second}
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

func PostForAgodaReservationsDetails(postUrl string, bookingId int64, cookie string, resBody interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	urlData := url.Values{}
	urlData.Set("BookingDetailList", fmt.Sprintf("[{\"BookingId\":\"%d\",\"EmailType\":\"1\"}]", bookingId))

	req, err := http.NewRequest("POST", postUrl, strings.NewReader(urlData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 120 * time.Second}
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
