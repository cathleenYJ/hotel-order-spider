package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gin/file"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ReservationsDB struct {
	Platform          string  `gorm:"uniqueIndex:platform_booking_id" json:"platform"`
	BookingId         string  `gorm:"uniqueIndex:platform_booking_id" json:"booking_id"`
	BookDate          string  `json:"book_date"`
	GuestName         string  `json:"guest_name"`
	NumOfGuests       int64   `json:"num_of_guests"`
	CheckInDate       string  `json:"check_in_date"`
	CheckOutDate      string  `json:"check_out_date"`
	Commission        float64 `json:"commission"`
	Price             float64 `json:"price"`
	Currency          string  `json:"currency"`
	ReservationStatus string  `json:"reservation_status"`
	NumOfRooms        int64   `json:"num_of_rooms"`
	GuestRequest      string  `json:"guest_request"`
	RoomNights        int64   `json:"room_nights"`
	HotelId           string  `json:"hotel_id"`
	Charge            string  `json:"charge"`
	RoomType          string  `json:"room_type"`
	ModifyAmt         string  `json:"modify_amt"`
}

type GetBookingReservationResponseBody struct {
	Data struct {
		Reservations []struct {
			GuestName         string `json:"guestName"`
			BookDate          string `json:"bookDate"`
			BookingId         int64  `json:"id"`
			CheckInDate       string `json:"checkin"`
			CheckOutDate      string `json:"checkout"`
			ReservationStatus string `json:"reservationStatus"`
			Commission        struct {
				Original struct {
					Amount    float64 `json:"amount"`
					Currency  string  `json:"currency"`
					Formatted string  `json:"formatted"`
				} `json:"original"`
			} `json:"commission"`
			Price struct {
				Currency          string  `json:"currency"`
				FormattedCurrency string  `json:"formatted"`
				Amount            float64 `json:"amount"`
			} `json:"price"`
			Occupancy struct {
				NumOfGuest int64 `json:"guests"`
				Adults     int64 `json:"adults"`
				Children   int64 `json:"children"`
			} `json:"occupancy"`
			Rooms []struct {
				Name     string `json:"name"`
				Quantity int64  `json:"quantity"`
			} `json:"rooms"`
		} `json:"reservations"`
	} `json:"data"`
}

func GetBooking(platform map[string]interface{}, platformName, period, dateFrom, dateTo string) {
	var result string
	var url string

	parse, ok := platform["parse"].(string)
	if !ok {
		fmt.Println("無法取得 parse")
	}

	cookie, ok := platform["cookie"].(string)
	if !ok {
		fmt.Println("無法取得 cookie")
	}

	session := GetBookingSessionID(cookie)
	fmt.Println(session)

	token, ok := platform["token"].(string)
	if !ok {
		fmt.Println("無法取得 token")
	}

	hotels, ok := platform["hotel"].([]interface{})
	if !ok || hotels == nil {
		fmt.Println("無法取得 hotel")
	}

	for _, hotelRaw := range hotels {
		hotel, ok := hotelRaw.(map[string]interface{})
		if !ok || hotel == nil {
			fmt.Println("無法取得 hotel")
			continue
		}

		hotelName, ok := hotel["name"].(string)
		if !ok {
			fmt.Println("無法取得 hotel name")
			continue
		}

		hotelId, ok := hotel["hotelid"].(string)
		if !ok {
			fmt.Println("無法取得 hotel id")
			continue
		}

		fmt.Printf("Hotel Name: %s, Hotel ID: %s\n", hotelName, hotelId)

		dateFromTime, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			fmt.Println("日期解析錯誤:", err)
			return
		}

		dateToTime, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			fmt.Println("日期解析錯誤:", err)
			return
		}

		var resultData []ReservationsDB

		if parse == "API" {
			for current := dateFromTime; current.Before(dateToTime) || current.Equal(dateToTime); current = current.AddDate(0, 0, 1) {
				currentDateString := current.Format("2006-01-02")
				fmt.Printf("處理日期：%s\n", currentDateString)
				url = fmt.Sprintf("https://admin.booking.com/fresa/extranet/reservations/retrieve_list_v2?lang=xt&ses=%s&hotel_id=%s&hotel_account_id=17606105&perpage=1000&page=1&date_type=departure&date_from=%s&date_to=%s&token=%s", session, hotelId, currentDateString, currentDateString, token)
				if err := DoRequestAndGetResponse("POST", url, http.NoBody, cookie, &result); err != nil {
					fmt.Println("DoRequestAndGetResponse failed!")
					fmt.Println("err", err)
					return
				}
				fmt.Println()

				// 解碼JSON
				var ordersData GetBookingReservationResponseBody
				err = json.Unmarshal([]byte(result), &ordersData)
				if err != nil {
					fmt.Println("JSON解碼錯誤:", err)
					return
				}

				var data ReservationsDB
				for _, reservation := range ordersData.Data.Reservations {
					data.BookingId = strconv.FormatInt(reservation.BookingId, 10)
					data.GuestName = reservation.GuestName

					data.CheckOutDate = reservation.CheckOutDate
					data.CheckInDate = reservation.CheckInDate

					checkInTime, err := time.Parse("2006-01-02", reservation.CheckInDate)
					if err != nil {
						fmt.Println("确认入住日期解析错误:", err)
						return
					}
					checkOutTime, err := time.Parse("2006-01-02", reservation.CheckOutDate)
					if err != nil {
						fmt.Println("确认退房日期解析错误:", err)
						return
					}
					duration := checkOutTime.Sub(checkInTime)
					days := int(duration.Hours() / 24)
					data.RoomNights = int64(days)

					data.Price = reservation.Price.Amount
					data.Commission = reservation.Commission.Original.Amount
					data.ReservationStatus = reservation.ReservationStatus
					data.Platform = platformName
					data.Currency = reservation.Price.Currency
					if reservation.Occupancy.NumOfGuest > 0 {
						data.NumOfGuests = reservation.Occupancy.NumOfGuest
					} else {
						data.NumOfGuests = reservation.Occupancy.Adults + reservation.Occupancy.Children
					}

					data.HotelId = hotelId

					if data.ReservationStatus == "ok" {
						data.ReservationStatus = "已成立"
						if data.Commission == 0 {
							data.ReservationStatus = "已取消"
						}
					} else if data.ReservationStatus == "no_show" {
						data.ReservationStatus = "已成立"
						if data.Commission == 0 {
							data.ReservationStatus = "已取消"
						}
						if data.Commission != 0 {
							data.ReservationStatus = "Chargeable no show"
						}
					} else if data.ReservationStatus == "cancelled_by_guest" || data.ReservationStatus == "cancelled_by_hotel" {
						data.ReservationStatus = "已取消"
						if data.Commission != 0 {
							data.ReservationStatus = "Chargeable cancellation"
						}
					}
					fmt.Println("data", data)
					resultData = append(resultData, data)
				}
			}
		} else if parse == "HTML" {
			/// 財務 訂單明細
			url = fmt.Sprintf("https://admin.booking.com/hotel/hoteladmin/extranet_ng/manage/finance_reservations.html?ses=%s&hotel_id=%s&period=%s&lang=xt", session, hotelId, period)
			if err := DoRequestAndGetResponse("GET", url, http.NoBody, cookie, &result); err != nil {
				fmt.Println("DoRequestAndGetResponse failed!")
				fmt.Println("err", err)
				return
			}

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
			if err != nil {
				log.Fatal(err)
				return
			}
			// 儲存已經存在的 BookingId
			existingBookingIds := make(map[string]bool)
			currency := ""
			doc.Find("#reservations tbody tr").Each(func(i int, s *goquery.Selection) {
				var data ReservationsDB
				var isDispute string
				s.Find("td").Each(func(j int, cell *goquery.Selection) {
					switch j {
					case 0:
						bookingId := cell.Find("span.visible-print").Text()
						data.BookingId = strings.TrimSpace(bookingId)

					case 1:
						data.GuestName = strings.TrimSpace(cell.Text())

					case 2:
						dateText := strings.TrimSpace(cell.Text())
						// 找到 "日" 的位置
						index := strings.Index(dateText, "日")
						if index != -1 {
							// 取得 "日" 之前的部分
							dateText = dateText[:index+3]
						}
						parsedTime, err := time.Parse("2006 年 1 月 2 日", dateText)
						if err != nil {
							fmt.Println("日期解析錯誤:", err)
							return
						}
						data.CheckInDate = parsedTime.Format("2006-01-02")

					case 3:
						dateText := strings.TrimSpace(cell.Text())
						// 找到 "日" 的位置
						index := strings.Index(dateText, "日")
						if index != -1 {
							// 取得 "日" 之前的部分
							dateText = dateText[:index+3]
						}
						parsedTime, err := time.Parse("2006 年 1 月 2 日", dateText)
						if err != nil {
							fmt.Println("日期解析錯誤:", err)
							return
						}
						data.CheckOutDate = parsedTime.Format("2006-01-02")

					case 4:
						roomNights, _ := strconv.Atoi(strings.TrimSpace(cell.Text()))
						data.RoomNights = int64(roomNights)

					case 6:
						data.ReservationStatus = strings.TrimSpace(cell.Text())
						// 使用 strings.Fields 分割字符串
						fields := strings.Fields(data.ReservationStatus)
						// 取得空格前的第一個元素
						if len(fields) > 0 {
							data.ReservationStatus = fields[0]
						}

					case 8:
						if strings.Contains(cell.Text(), "TWD") {
							currency = "TWD"
							priceStr := strings.Replace(cell.Text(), "TWD ", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.TrimSpace(priceStr)
							// 使用 strings.Fields 分割字符串
							fields := strings.Fields(priceStr)
							// 取得空格前的第一個元素
							if len(fields) > 0 {
								priceStr = fields[0]
							}
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
								return
							}
							data.Price = price
						} else if strings.Contains(cell.Text(), "US$") {
							currency = "USD"
							priceStr := strings.Replace(cell.Text(), "US$", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.TrimSpace(priceStr)
							// 使用 strings.Fields 分割字符串
							fields := strings.Fields(priceStr)
							// 取得空格前的第一個元素
							if len(fields) > 0 {
								priceStr = fields[0]
							}
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
								return
							}
							data.Price = price
						} else if strings.Contains(cell.Text(), "¥") {
							currency = "JPY"
							priceStr := strings.Replace(cell.Text(), "¥", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.TrimSpace(priceStr)
							// 使用 strings.Fields 分割字符串
							fields := strings.Fields(priceStr)
							// 取得空格前的第一個元素
							if len(fields) > 0 {
								priceStr = fields[0]
							}
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
								return
							}
							data.Price = price
						}

					case 9:
						priceStr := cell.Text()
						if strings.Contains(priceStr, "TWD") {
							priceStr = strings.Replace(priceStr, "TWD ", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.Replace(priceStr, " ", "", -1)
							priceStr = strings.TrimSpace(priceStr)
							fields := strings.Fields(priceStr)
							if len(fields) > 0 {
								priceStr = fields[0]
							}
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
								return
							}
							data.Commission = price
						} else if strings.Contains(cell.Text(), "US$") {
							priceStr := strings.Replace(cell.Text(), "US$", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.Replace(priceStr, " ", "", -1)
							priceStr = strings.TrimSpace(priceStr)
							fields := strings.Fields(priceStr)
							if len(fields) > 0 {
								priceStr = fields[0]
							}
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
								return
							}
							data.Commission = price
						} else if strings.Contains(cell.Text(), "¥") {
							priceStr := strings.Replace(cell.Text(), "¥", "", -1)
							priceStr = strings.Replace(priceStr, ",", "", -1)
							priceStr = strings.Replace(priceStr, " ", "", -1)
							priceStr = strings.TrimSpace(priceStr)
							fields := strings.Fields(priceStr)
							if len(fields) > 0 {
								priceStr = fields[0]
							}
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
								return
							}
							data.Commission = price
						}

					case 10:
						if cell.Find(".glyphicon-ok").Length() > 0 {
							dispute := cell.Find(".glyphicon-ok")
							dispute.Each(func(i int, s *goquery.Selection) {
								isDispute = "dispute"
								fmt.Println("dispute", data.BookingId)
							})
						} else {
							isDispute = ""
						}

					case 11:
						charge := cell.Text()
						charge = strings.Replace(charge, " ", "", -1)
						data.Charge = charge
					}
					data.Platform = platformName
					data.HotelId = hotelId
					data.Currency = currency
					// 檢查 BookingId 是否為空或是已經存在 existingBookingIds 中，如果是，就不加入 resultData
					if data.BookingId != "" && data.ReservationStatus != "" && data.Charge != "" {
						if !existingBookingIds[data.BookingId] {
							if data.ReservationStatus == "已入住" {
								data.ReservationStatus = "已取消"
								if data.Commission != 0 {
									data.ReservationStatus = "已成立"
								}
							} else if data.ReservationStatus == "取消" {
								data.ReservationStatus = "已取消"
								if data.Commission != 0 {
									data.ReservationStatus = "Chargeable cancellation"
								}
							} else if data.ReservationStatus == "未如期入住" {
								data.ReservationStatus = "已取消"
								if data.Commission != 0 {
									data.ReservationStatus = "Chargeable no show"
								}
							}

							if isDispute == "dispute" {
								// data.Price = 0
								// data.Commission = 0
								data.ModifyAmt = "discount"
							}

							resultData = append(resultData, data)
							file.AppendToFile("hotel_orders.txt", data.BookingId+"\t"+data.GuestName+"\t"+data.CheckInDate+"\t"+data.CheckOutDate+"\t"+strconv.FormatInt(data.RoomNights, 10)+"\t"+data.ReservationStatus+"\t"+strconv.FormatFloat(data.Price, 'f', 2, 64)+"\t"+strconv.FormatFloat(data.Commission, 'f', 2, 64)+"\n")
							// 將目前的 BookingId 添加到 existingBookingIds 中
							existingBookingIds[data.BookingId] = true
						}
					}
				})
			})
			fmt.Println("resultdata", resultData)
		}

		// 將 data 轉換為 JSON 格式
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
}

func GetBookingSessionID(cookie string) string {
	const url = "https://admin.booking.com/hotel/hoteladmin/groups/home/index.html"

	var result string
	if err := DoRequestAndGetResponse("GET", url, http.NoBody, cookie, &result); err != nil {
		return "failed"
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		log.Fatal(err)
		return "failed"
	}

	var session = ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		_, ok := s.Attr("data-capla-application-context")
		if ok {
			text := s.Text()
			re := regexp.MustCompile(`ses=(.*)","currency`)
			match := re.FindStringSubmatch(text)
			if len(match) > 1 {
				session = match[1]
				fmt.Println("match found -", match[1])
			} else {
				fmt.Println("match not found")
			}
		}
	})

	return session
}

func DoRequestAndGetResponse(method, postUrl string, reqBody io.Reader, cookie string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", cookie)
	switch resBody := resBody.(type) {
	case *string:
		fmt.Println("string")
		fmt.Println(resBody)

		req.Header.Set("Content-Type", "application/json")
	default:
		fmt.Println("not string")
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 100 * time.Second}
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
