package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GetHostelworld(platform map[string]interface{}, platformName, dateFrom, dateTo string) {
	var url string
	var result string

	if hotelsRaw, ok := platform["hotel"]; ok {
		hotels, ok := hotelsRaw.([]interface{})
		if !ok || hotels == nil {
			fmt.Println("無法取得 hotel")
		}

		for _, hotelRaw := range hotels {
			hotel, _ := hotelRaw.(map[string]interface{})
			hotelName, _ := hotel["name"].(string)
			cookie, _ := hotel["cookie"].(string)

			url = fmt.Sprintf("https://inbox.hostelworld.com/booking/search/date?DateType=departuredate&dateFrom=%s&dateTo=%s", dateFrom, dateTo)
			if err := DoRequestAndGetResponse("GET", url, http.NoBody, cookie, &result); err != nil {
				fmt.Println("DoRequestAndGetResponse failed!")
				fmt.Println("err", err)
				return
			}

			var resultData []ReservationsDB
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
			if err != nil {
				log.Fatal(err)
				return
			}

			// 儲存已經存在的 BookingId
			existingBookingIds := make(map[string]bool)

			fmt.Printf("Hotel Name: %s\n", hotelName)

			doc.Find(".bookings_container tbody tr").Each(func(i int, s *goquery.Selection) {
				var data ReservationsDB
				s.Find("td").Each(func(j int, cell *goquery.Selection) {
					switch j {
					case 0:
						linkText := strings.TrimSpace(cell.Find("a").Text())
						data.BookingId = linkText
						fmt.Println(data.BookingId)
					case 1:
						data.GuestName = strings.TrimSpace(cell.Text())
						fmt.Println(data.GuestName)
					case 2:
						cleanedDate := cleanDateString(cell.Text())
						parsedTime, err := time.Parse("2 Jan '06", cleanedDate)
						if err != nil {
							fmt.Println("日期解析失败:", err)
							return
						}
						data.CheckOutDate = parsedTime.Format("2006-01-02")
						fmt.Println(data.CheckOutDate)

					case 3:
						data.Price, err = strconv.ParseFloat(cell.Text(), 64)
						fmt.Println(data.Price)
					case 4:
						cleanedDate := cleanDateString(cell.Text())
						parsedTime, err := time.Parse("2 Jan '06", cleanedDate)
						if err != nil {
							fmt.Println("日期解析失败:", err)
							return
						}
						data.BookDate = parsedTime.Format("2006-01-02")
						fmt.Println(data.BookDate)
					case 5:
						roomNights, _ := strconv.Atoi(strings.TrimSpace(cell.Text()))
						data.RoomNights = int64(roomNights)
						fmt.Println(data.RoomNights)
					case 6:
						data.GuestRequest = cell.Text()
						fmt.Println(data.GuestRequest)
					case 7:
						roomType := strings.Replace(cell.Text(), " ", "", -1)
						data.RoomType = roomType
						fmt.Println(data.RoomType)
					}

					data.ReservationStatus = "已成立"
					fmt.Println(data.ReservationStatus)

					data.Platform = platformName
					fmt.Println(data.Platform)

					hotelIdFrombookingId := strings.Split(data.BookingId, "-")
					if len(hotelIdFrombookingId) > 0 {
						data.HotelId = hotelIdFrombookingId[0]
						fmt.Println(data.HotelId)
					}

					data.Currency = "TWD"
					fmt.Println(data.Currency)

					// 檢查 BookingId 是否为空或已存在于 existingBookingIds 中，如果是，就不加入 resultData
					if data.BookingId != "" && data.Price != 0 && data.RoomType != "" {
						if !existingBookingIds[data.BookingId] {
							resultData = append(resultData, data)
							existingBookingIds[data.BookingId] = true

						}
					}
				})
			})
			fmt.Println("resultData", resultData)

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
}

// 清理日期字串中的 "th", "st", "nd", "rd" 和 0
func cleanDateString(date string) string {
	// 替換 "th", "st", "nd", "rd"
	re := regexp.MustCompile(`(\d+)(st|nd|rd|th)`)
	cleanedDate := re.ReplaceAllString(date, "$1")

	// 去除 0
	cleanedDate = strings.TrimLeft(cleanedDate, "0")

	return cleanedDate
}
