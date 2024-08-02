package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type GetTripVCCOrderResponseBody struct {
	Data []struct {
		HotelId            int64   `json:"hotelId"`
		OrderId            int64   `json:"orderId"`
		CardInfoDataID     string  `json:"cardInfoDataID"`
		GuestName          string  `json:"guestName"`
		CheckInDate        string  `json:"checkInDate"`
		CheckOutDate       string  `json:"checkOutDate"`
		EffectiveDate      *string `json:"effectiveDate"`
		Charge             float64 `json:"charge"`
		Refund             float64 `json:"refund"`
		Balance            float64 `json:"balance"`
		RefundDeadline     string  `json:"refundDeadline"`
		Currency           string  `json:"currency"`
		CardCurrency       string  `json:"cardCurrency"`
		AdjustAmountStatus string  `json:"adjustAmountStatus"`
		OverchargeStatus   string  `json:"overchargeStatus"`
	} `json:"data"`
}

type TripWithdrawnPrePaidData struct {
	Data struct {
		PageIndex   int                                        `json:"pageIndex"`
		TotalPage   int                                        `json:"totalPage"`
		PageSize    int                                        `json:"pageSize"`
		TotalRecord int                                        `json:"totalRecord"`
		DataList    []GetTripWithdrawnPrePaidOrderResponseBody `json:"dataList"`
	} `json:"data"`
}

type GetTripWithdrawnPrePaidOrderResponseBody struct {
	OrderId         string  `json:"orderId"`
	HotelId         int     `json:"hotelId"`
	ClientName      string  `json:"clientName"`
	Eta             string  `json:"eta"`
	Etd             string  `json:"etd"`
	Quantity        int     `json:"quantity"`
	Currency        string  `json:"currency"`
	SettlementPrice float64 `json:"settlementPrice"`
}

type TripPrePaidData struct {
	Data struct {
		PageIndex   int                               `json:"pageIndex"`
		TotalPage   int                               `json:"totalPage"`
		PageSize    int                               `json:"pageSize"`
		TotalRecord int                               `json:"totalRecord"`
		DataList    []GetTripPrePaidOrderResponseBody `json:"dataList"`
	} `json:"data"`
}

type GetTripPrePaidOrderResponseBody struct {
	OrderId    string  `json:"orderId"`
	HotelId    int     `json:"hotelId"`
	ClientName string  `json:"clientName"`
	Eta        string  `json:"etaStr"`
	Etd        string  `json:"etdStr"`
	Quantity   int     `json:"quantity"`
	Currency   string  `json:"currency"`
	OrderPrice float64 `json:"orderPrice"`
}

type TripPaidData struct {
	Data struct {
		PageIndex   int                            `json:"pageIndex"`
		TotalPage   int                            `json:"totalPage"`
		PageSize    int                            `json:"pageSize"`
		TotalRecord int                            `json:"totalRecord"`
		DataList    []GetTripPaidOrderResponseBody `json:"dataList"`
	} `json:"data"`
}

type GetTripPaidOrderResponseBody struct {
	OrderId           string  `json:"orderId"`
	HotelId           int     `json:"hotelId"`
	ClientName        string  `json:"clientName"`
	ConfirmETA        string  `json:"confirmETA"`
	ConfirmETD        string  `json:"confirmETD"`
	Currency          string  `json:"currency"`
	ConfirmCostAmount float64 `json:"confirmCostAmount"`
}

type TripEBKData struct {
	Data struct {
		Orders []GetTripEBKOrderResponseBody `json:"orders"`
	} `json:"data"`
}

type GetTripEBKOrderResponseBody struct {
	OrderId            string  `json:"orderID"`
	HotelId            int     `json:"hotel"`
	ClientName         string  `json:"clientNameDomestic"`
	Arrival            string  `json:"Arrival"`
	Departure          string  `json:"Departure"`
	Currency           string  `json:"currency"`
	OrderStatusDisplay string  `json:"orderStatusDisplay"`
	Amount             string  `json:"amount"`
	RoomNights         float64 `json:"LiveDays"`
}

func GetCtrip(platform map[string]interface{}, platformName, accountName, dateFrom, dateTo string) {

	var url string
	var result string
	var resultChangeHotel string

	cookie, _ := platform["cookie"].(string)
	hotelsRaw, ok := platform["hotel"]
	if ok {
		hotels, ok := hotelsRaw.([]interface{})
		if ok && hotels != nil {
			for _, hotelRaw := range hotels {
				hotel, _ := hotelRaw.(map[string]interface{})
				hotelName, _ := hotel["name"].(string)
				fmt.Printf("hotelName: %s", hotelName)

				hotelid, _ := hotel["hotelid"].(string)
				fmt.Printf("hotelid: %s", hotelid)

				hotelid_pre, _ := hotel["hotelid_pre"].(string)
				fmt.Printf("hotelid_pre: %s", hotelid_pre)

				masterhotelid, _ := hotel["masterhotelid"].(string)
				fmt.Printf("masterhotelid: %s", masterhotelid)

				batchid, _ := hotel["batchid"].(string)
				fmt.Printf("batchid: %s", batchid)

				token, _ := hotel["token"].(string)
				fmt.Printf("token: %s", token)

				if accountName != "上海雀の花园民宿" && accountName != "豐豐民宿" {
					// VCC
					if hotelid_pre != "" {
						fmt.Println()
						fmt.Println("VCC")

						changeHotelUrl := "https://ebooking.ctrip.com/restapi/soa2/23270/changeHotel?_fxpcqlniredt=09031138411810763313&x-traceID=09031138411810763313-1691983521082-8683634"
						reqBodyStrChangeHotel := fmt.Sprintf("{\"reqHead\":{\"host\":\"ebooking.ctrip.com\",\"locale\":\"en-US\",\"release\":\"\",\"client\":{\"deviceType\":\"PC\",\"os\":\"Mac\",\"osVersion\":\"\",\"clientId\":\"09031138411810763313\",\"screenWidth\":1440,\"screenHeight\":900,\"isIn\":{\"ie\":false,\"chrome\":false,\"wechat\":false,\"firefox\":false,\"ios\":false,\"android\":false},\"isModernBrowser\":true,\"browser\":\"Safari\",\"browserVersion\":\"\",\"platform\":\"\",\"technology\":\"\"},\"ubt\":{\"pageid\":10650072645,\"pvid\":20,\"sid\":6,\"vid\":\"1690777948792.2i9kdb\"},\"gps\":{\"coord\":\"\",\"lat\":\"\",\"lng\":\"\",\"cid\":0,\"cnm\":\"\"},\"protocal\":\"https:\"},\"hotelId\":%s,\"masterHotelId\":%s,\"head\":{\"cid\":\"09031138411810763313\",\"ctok\":\"\",\"cver\":\"1.0\",\"lang\":\"01\",\"sid\":\"8888\",\"syscode\":\"09\",\"auth\":\"\",\"xsid\":\"\",\"extension\":[]}}", hotelid_pre, masterhotelid)
						jsonReqBodyChangeHotel := []byte(reqBodyStrChangeHotel)

						if err := DoRequestAndGetResponse("POST", changeHotelUrl, bytes.NewBuffer(jsonReqBodyChangeHotel), cookie, &resultChangeHotel); err != nil {
							fmt.Println("DoRequestAndGetResponse failed!")
							fmt.Println("err", err)
							return
						}
						fmt.Println("change hotel success!")

						url = "https://ebooking.ctrip.com/ebkfinance/vcc/queryHotelVCCOrderInfo"
						reqBodyStr := fmt.Sprintf("{\"dateType\":\"DepartureDate\",\"dateStart\":\"%s\",\"dateEnd\":\"%s\",\"beginCheckInDate\":\"\",\"endCheckInDate\":\"\",\"beginCheckOutDate\":\"%s\",\"endCheckOutDate\":\"%s\",\"hotelId\":%v,\"orderType\":3}", dateFrom, dateTo, dateFrom, dateTo, hotelid_pre)
						jsonReqBody := []byte(reqBodyStr)

						fmt.Println()

						if err := DoRequestAndGetResponse("POST", url, bytes.NewBuffer(jsonReqBody), cookie, &result); err != nil {
							fmt.Println("DoRequestAndGetResponse failed!")
							fmt.Println("err", err)
							return
						}

						var resultData []ReservationsDB

						// 解碼JSON
						var ordersData GetTripVCCOrderResponseBody
						err := json.Unmarshal([]byte(result), &ordersData)
						if err != nil {
							fmt.Println("JSON解碼錯誤:", err)
							return
						}

						priceMap := make(map[string]float64)

						for _, reservation := range ordersData.Data {
							bookingID := strconv.FormatInt(reservation.OrderId, 10)
							if _, ok := priceMap[bookingID]; !ok {
								priceMap[bookingID] = 0
							}
							priceMap[bookingID] += reservation.Charge
						}

						var data ReservationsDB
						for _, reservation := range ordersData.Data {
							bookingID := strconv.FormatInt(reservation.OrderId, 10)
							if price, ok := priceMap[bookingID]; ok {
								data.BookingId = bookingID
								data.GuestName = reservation.GuestName

								arrivalTime, err := time.Parse("2006-01-02 15:04:05", reservation.CheckInDate)
								if err != nil {
									fmt.Println("Error parsing arrival time:", err)
								}

								departureTime, err := time.Parse("2006-01-02 15:04:05", reservation.CheckOutDate)
								if err != nil {
									fmt.Println("Error parsing arrival time:", err)
								}
								checkOutTime := departureTime
								checkInTime := arrivalTime
								data.CheckOutDate = checkOutTime.Format("2006-01-02")
								data.CheckInDate = checkInTime.Format("2006-01-02")

								duration := checkOutTime.Sub(checkInTime)
								roomNights := int64(duration.Hours()/24) + 1

								data.RoomNights = roomNights

								data.Price = price

								if reservation.EffectiveDate == nil {
									fmt.Println("reservation.EffectiveDate", reservation.EffectiveDate)
									data.ReservationStatus = "已取消"
									fmt.Println("data.ReservationStatus", data.ReservationStatus)
								} else {
									data.ReservationStatus = "已成立"
								}

								data.Platform = platformName
								data.Currency = reservation.Currency
								data.HotelId = hotelid_pre

								resultData = append(resultData, data)
							}
						}
						fmt.Println("VCC resultdata", resultData)
						time.Sleep(5 * time.Second)

						if len(resultData) != 0 {
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
						} else if len(resultData) == 0 {
							//預付-已提現
							if hotelid_pre != "" {
								fmt.Println()
								fmt.Println("預付-已提現")

								changeHotelUrl := "https://ebooking.ctrip.com/restapi/soa2/23270/changeHotel?_fxpcqlniredt=09031138411810763313&x-traceID=09031138411810763313-1691983521082-8683634"
								reqBodyStrChangeHotel := fmt.Sprintf("{\"reqHead\":{\"host\":\"ebooking.ctrip.com\",\"locale\":\"en-US\",\"release\":\"\",\"client\":{\"deviceType\":\"PC\",\"os\":\"Mac\",\"osVersion\":\"\",\"clientId\":\"09031138411810763313\",\"screenWidth\":1440,\"screenHeight\":900,\"isIn\":{\"ie\":false,\"chrome\":false,\"wechat\":false,\"firefox\":false,\"ios\":false,\"android\":false},\"isModernBrowser\":true,\"browser\":\"Safari\",\"browserVersion\":\"\",\"platform\":\"\",\"technology\":\"\"},\"ubt\":{\"pageid\":10650072645,\"pvid\":20,\"sid\":6,\"vid\":\"1690777948792.2i9kdb\"},\"gps\":{\"coord\":\"\",\"lat\":\"\",\"lng\":\"\",\"cid\":0,\"cnm\":\"\"},\"protocal\":\"https:\"},\"hotelId\":%s,\"masterHotelId\":%s,\"head\":{\"cid\":\"09031138411810763313\",\"ctok\":\"\",\"cver\":\"1.0\",\"lang\":\"01\",\"sid\":\"8888\",\"syscode\":\"09\",\"auth\":\"\",\"xsid\":\"\",\"extension\":[]}}", hotelid_pre, masterhotelid)
								jsonReqBodyChangeHotel := []byte(reqBodyStrChangeHotel)

								if err := DoRequestAndGetResponse("POST", changeHotelUrl, bytes.NewBuffer(jsonReqBodyChangeHotel), cookie, &resultChangeHotel); err != nil {
									fmt.Println("DoRequestAndGetResponse failed!")
									fmt.Println("err", err)
									return
								}
								fmt.Println("change hotel success!")

								url = "https://ebooking.ctrip.com/restapi/soa2/29140/getPrepayOrders?_fxpcqlniredt=09031111118208455165&x-traceID=09031111118208455165-1706177694881-1207637"
								reqBodyStr := fmt.Sprintf("{\"hotelId\":%v,\"orderId\":\"\",\"startETA\":\"\",\"endETA\":\"\",\"startETD\":\"%s\",\"endETD\":\"%s\",\"startInputTime\":\"\",\"endInputTime\":\"\",\"startFinishTime\":\"\",\"endFinishTime\":\"\",\"isShortRent\":false,\"tabType\":2,\"pageIndex\":1,\"pageSize\":500}", hotelid_pre, dateFrom, dateTo)
								jsonReqBody := []byte(reqBodyStr)

								fmt.Println()

								if err := DoRequestAndGetResponse("POST", url, bytes.NewBuffer(jsonReqBody), cookie, &result); err != nil {
									fmt.Println("DoRequestAndGetResponse failed!")
									fmt.Println("err", err)
									return
								}

								var resultData []ReservationsDB

								// 解碼JSON
								var ordersData TripWithdrawnPrePaidData
								err := json.Unmarshal([]byte(result), &ordersData)
								if err != nil {
									fmt.Println("JSON解碼錯誤:", err)
									return
								}

								priceMap := make(map[string]float64)

								for _, reservation := range ordersData.Data.DataList {
									bookingID := reservation.OrderId
									if _, ok := priceMap[bookingID]; !ok {
										priceMap[bookingID] = 0
									}
									priceMap[bookingID] += reservation.SettlementPrice
								}

								var data ReservationsDB
								for _, reservation := range ordersData.Data.DataList {
									bookingID := reservation.OrderId
									if price, ok := priceMap[bookingID]; ok {
										data.BookingId = bookingID
										data.GuestName = reservation.ClientName

										data.CheckOutDate = reservation.Etd
										data.CheckInDate = reservation.Eta

										data.RoomNights = int64(reservation.Quantity)

										data.Price = price
										data.ReservationStatus = "已成立"
										if data.Price == 0 {
											data.ReservationStatus = "已取消"
										}
										data.Platform = "Ctrip"
										data.Currency = reservation.Currency
										data.HotelId = hotelid_pre

										resultData = append(resultData, data)
									}
								}
								fmt.Println("預付-已提現 resultdata", resultData)
								time.Sleep(5 * time.Second)

								if len(resultData) != 0 {
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

							//預付-待提現
							if hotelid_pre != "" {
								fmt.Println()
								fmt.Println("預付-待提現")

								changeHotelUrl := "https://ebooking.ctrip.com/restapi/soa2/29140/getPrepayOrders?_fxpcqlniredt=09031111118208455165&x-traceID=09031111118208455165-1706770959306-5924108"
								reqBodyStrChangeHotel := fmt.Sprintf("{\"reqHead\":{\"host\":\"ebooking.ctrip.com\",\"locale\":\"en-US\",\"release\":\"\",\"client\":{\"deviceType\":\"PC\",\"os\":\"Mac\",\"osVersion\":\"\",\"clientId\":\"09031138411810763313\",\"screenWidth\":1440,\"screenHeight\":900,\"isIn\":{\"ie\":false,\"chrome\":false,\"wechat\":false,\"firefox\":false,\"ios\":false,\"android\":false},\"isModernBrowser\":true,\"browser\":\"Safari\",\"browserVersion\":\"\",\"platform\":\"\",\"technology\":\"\"},\"ubt\":{\"pageid\":10650072645,\"pvid\":20,\"sid\":6,\"vid\":\"1690777948792.2i9kdb\"},\"gps\":{\"coord\":\"\",\"lat\":\"\",\"lng\":\"\",\"cid\":0,\"cnm\":\"\"},\"protocal\":\"https:\"},\"hotelId\":%s,\"masterHotelId\":%s,\"head\":{\"cid\":\"09031138411810763313\",\"ctok\":\"\",\"cver\":\"1.0\",\"lang\":\"01\",\"sid\":\"8888\",\"syscode\":\"09\",\"auth\":\"\",\"xsid\":\"\",\"extension\":[]}}", hotelid_pre, masterhotelid)
								jsonReqBodyChangeHotel := []byte(reqBodyStrChangeHotel)

								if err := DoRequestAndGetResponse("POST", changeHotelUrl, bytes.NewBuffer(jsonReqBodyChangeHotel), cookie, &resultChangeHotel); err != nil {
									fmt.Println("DoRequestAndGetResponse failed!")
									fmt.Println("err", err)
									return
								}
								fmt.Println("change hotel success!")

								url = "https://ebooking.ctrip.com/restapi/soa2/29140/getPrepayUncollectedOrders?_fxpcqlniredt=09031043311840621561&x-traceID=09031043311840621561-1719809437718-6887579"
								reqBodyStr := fmt.Sprintf("{\"hotelId\":%v,\"orderId\":\"\",\"startETA\":\"\",\"endETA\":\"\",\"startETD\":\"%s\",\"endETD\":\"%s\",\"startInputTime\":\"\",\"endInputTime\":\"\",\"startFinishTime\":\"\",\"endFinishTime\":\"\",\"isShortRent\":false,\"tabType\":1,\"pageIndex\":1,\"pageSize\":500}", hotelid_pre, dateFrom, dateTo)
								jsonReqBody := []byte(reqBodyStr)

								fmt.Println()

								if err := DoRequestAndGetResponse("POST", url, bytes.NewBuffer(jsonReqBody), cookie, &result); err != nil {
									fmt.Println("DoRequestAndGetResponse failed!")
									fmt.Println("err", err)
									return
								}

								var resultData []ReservationsDB

								// 解碼JSON
								var ordersData TripPrePaidData
								err := json.Unmarshal([]byte(result), &ordersData)
								if err != nil {
									fmt.Println("JSON解碼錯誤:", err)
									return
								}

								priceMap := make(map[string]float64)

								for _, reservation := range ordersData.Data.DataList {
									bookingID := reservation.OrderId
									if _, ok := priceMap[bookingID]; !ok {
										priceMap[bookingID] = 0
									}
									priceMap[bookingID] += reservation.OrderPrice
								}

								var data ReservationsDB
								for _, reservation := range ordersData.Data.DataList {
									bookingID := reservation.OrderId
									if price, ok := priceMap[bookingID]; ok {
										data.BookingId = bookingID
										data.GuestName = reservation.ClientName

										data.CheckOutDate = reservation.Etd
										data.CheckInDate = reservation.Eta

										data.RoomNights = int64(reservation.Quantity)

										data.Price = price
										data.ReservationStatus = "已成立"
										if data.Price == 0 {
											data.ReservationStatus = "已取消"
										}

										data.Platform = "Ctrip"
										data.Currency = reservation.Currency
										data.HotelId = hotelid_pre

										resultData = append(resultData, data)
									}
								}
								fmt.Println("預付-待提現 resultdata", resultData)
								time.Sleep(5 * time.Second)

								if len(resultData) != 0 {
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
					}

					//現付
					if hotelid != "" && batchid != "" && token != "" {
						fmt.Println()
						fmt.Println("現付")

						changeHotelUrl := "https://ebooking.ctrip.com/restapi/soa2/23270/changeHotel?_fxpcqlniredt=09031138411810763313&x-traceID=09031138411810763313-1691983521082-8683634"
						reqBodyStrChangeHotel := fmt.Sprintf("{\"reqHead\":{\"host\":\"ebooking.ctrip.com\",\"locale\":\"en-US\",\"release\":\"\",\"client\":{\"deviceType\":\"PC\",\"os\":\"Mac\",\"osVersion\":\"\",\"clientId\":\"09031138411810763313\",\"screenWidth\":1440,\"screenHeight\":900,\"isIn\":{\"ie\":false,\"chrome\":false,\"wechat\":false,\"firefox\":false,\"ios\":false,\"android\":false},\"isModernBrowser\":true,\"browser\":\"Safari\",\"browserVersion\":\"\",\"platform\":\"\",\"technology\":\"\"},\"ubt\":{\"pageid\":10650072645,\"pvid\":20,\"sid\":6,\"vid\":\"1690777948792.2i9kdb\"},\"gps\":{\"coord\":\"\",\"lat\":\"\",\"lng\":\"\",\"cid\":0,\"cnm\":\"\"},\"protocal\":\"https:\"},\"hotelId\":%s,\"masterHotelId\":%s,\"head\":{\"cid\":\"09031138411810763313\",\"ctok\":\"\",\"cver\":\"1.0\",\"lang\":\"01\",\"sid\":\"8888\",\"syscode\":\"09\",\"auth\":\"\",\"xsid\":\"\",\"extension\":[]}}", hotelid, masterhotelid)
						jsonReqBodyChangeHotel := []byte(reqBodyStrChangeHotel)

						if err := DoRequestAndGetResponse("POST", changeHotelUrl, bytes.NewBuffer(jsonReqBodyChangeHotel), cookie, &resultChangeHotel); err != nil {
							fmt.Println("DoRequestAndGetResponse failed!")
							fmt.Println("err", err)
							return
						}
						fmt.Println("change hotel success!")

						pageSize := 100
						pageIndex := 1

						url = "https://ebooking.ctrip.com/ebkfinance/fgpay/getAllBatchOrderDetails"
						reqBodyStr := fmt.Sprintf("{\"batchId\":\"%s\",\"hotelId\":\"%s\",\"token\":\"%s\",\"pageIndex\":%v,\"pageSize\":%v,\"searchResultFromBill\":true,\"checkInStartTime\":\"\",\"checkInEndTime\":\"\",\"checkOutStartTime\":\"\",\"checkOutEndTime\":\"\"}", batchid, hotelid, token, pageIndex, pageSize)
						jsonReqBody := []byte(reqBodyStr)

						fmt.Println()

						if err := DoRequestAndGetResponse("POST", url, bytes.NewBuffer(jsonReqBody), cookie, &result); err != nil {
							fmt.Println("DoRequestAndGetResponse failed!")
							fmt.Println("err", err)
							return
						}

						var resultData []ReservationsDB

						// 解碼JSON
						var ordersData TripPaidData
						err := json.Unmarshal([]byte(result), &ordersData)
						if err != nil {
							fmt.Println("JSON解碼錯誤:", err)
							return
						}

						priceMap := make(map[string]float64)

						for _, reservation := range ordersData.Data.DataList {
							bookingID := reservation.OrderId
							if _, ok := priceMap[bookingID]; !ok {
								priceMap[bookingID] = 0
							}
							priceMap[bookingID] += reservation.ConfirmCostAmount
						}

						var data ReservationsDB
						for _, reservation := range ordersData.Data.DataList {
							bookingID := reservation.OrderId
							if price, ok := priceMap[bookingID]; ok {
								data.BookingId = bookingID
								data.GuestName = reservation.ClientName

								data.CheckOutDate = reservation.ConfirmETD
								data.CheckInDate = reservation.ConfirmETA

								etaTime, err := time.Parse("2006-01-02", reservation.ConfirmETA)
								if err != nil {
									fmt.Println("确认入住日期解析错误:", err)
									return
								}
								etdTime, err := time.Parse("2006-01-02", reservation.ConfirmETD)
								if err != nil {
									fmt.Println("确认退房日期解析错误:", err)
									return
								}
								duration := etdTime.Sub(etaTime)
								days := int(duration.Hours() / 24)
								data.RoomNights = int64(days)

								data.Price = price
								data.ReservationStatus = "已成立"
								data.Platform = "Ctrip"
								data.Currency = reservation.Currency
								data.HotelId = hotelid

								resultData = append(resultData, data)
							}
						}
						fmt.Println("現付 resultData:", resultData)
						time.Sleep(5 * time.Second)

						if len(resultData) != 0 {
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

				//EBK
				if accountName == "上海雀の花园民宿" || accountName == "豐豐民宿" {

					pageIndex := 1

					for {
						url = "https://ebooking.ctrip.com/ebkorderv2/api/order/domestic/allOrderList"
						reqBodyStr := fmt.Sprintf("{\"customFilter\":\"None\",\"dateType\":\"DepartureDate\",\"dateStart\":\"%s\",\"dateEnd\":\"%s\",\"orderType\":\"ALL\",\"orderStatus\":\"0\",\"orderID\":\"\",\"clientName\":\"\",\"bookingNo\":\"\",\"roomName\":\"\",\"confirmName\":\"\",\"isGuarantee\":false,\"isPP\":false,\"isFG\":false,\"isUrgent\":false,\"isCreditOrder\":false,\"receiveType\":\"\",\"isHoldRoom\":false,\"isFreeSale\":false,\"sourceType\":\"Ebooking\",\"AllinanceName\":\"\",\"UnBookingInvoice\":false,\"IsBookingInvoice\":false,\"GuestComplaintsStatus\":null,\"OrderFAQStatus\":null,\"isUnionMember\":false,\"showSupplierOrderCount\":false,\"showCTOrderCount\":false,\"showOfficialOrderCount\":false,\"ClientDateTime\":\"2024-02-01 15:41:00\",\"DayOffSetType\":-99,\"RoomTicketOrder\":null,\"IsImComfirmOrder\":null,\"IsFreeRoomOrder\":false,\"IsXcsmz\":null,\"pageInfo\":{\"pageIndex\":%v,\"orderBy\":\"FormDate\"},\"searchAllHotel\":false,\"loadFirstDetail\":true,\"pageId\":\"10650061100\",\"spiderkey\":\"12d096697493c0d0080fbd7aec6fe46da9358c9eaef4713cacaa6385f0fa91f2\",\"hoteluuidkeys\":\"\"}", dateFrom, dateTo, pageIndex)
						jsonReqBody := []byte(reqBodyStr)

						fmt.Println()

						if err := DoRequestAndGetResponse("POST", url, bytes.NewBuffer(jsonReqBody), cookie, &result); err != nil {
							fmt.Println("DoRequestAndGetResponse failed!")
							fmt.Println("err", err)
							return
						}

						var resultData []ReservationsDB

						// 解碼JSON
						var ordersData TripEBKData
						err := json.Unmarshal([]byte(result), &ordersData)
						if err != nil {
							fmt.Println("JSON解碼錯誤:", err)
							return
						}

						var data ReservationsDB
						for _, reservation := range ordersData.Data.Orders {
							data.BookingId = reservation.OrderId
							data.GuestName = reservation.ClientName

							data.CheckOutDate = reservation.Departure
							data.CheckInDate = reservation.Arrival
							data.RoomNights = int64(reservation.RoomNights)

							price, err := strconv.ParseFloat(reservation.Amount, 64)
							if err != nil {
								fmt.Println("Error converting string to float64:", err)
								return
							}
							data.Price = price

							data.ReservationStatus = reservation.OrderStatusDisplay
							if data.ReservationStatus == "已過離店日期" {
								data.ReservationStatus = "已成立"

								if data.Price == 0 {
									data.ReservationStatus = "已取消"
								}
							}
							if data.ReservationStatus == "已改订" && data.Price != 0 {
								data.ReservationStatus = "已成立"
							}

							data.Platform = "Ctrip"
							data.Currency = "RMB"
							data.HotelId = hotelid_pre

							resultData = append(resultData, data)
						}

						pageIndex++
						time.Sleep(5 * time.Second)
						fmt.Println("resultdata", resultData)

						if len(resultData) != 0 {
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
						if len(ordersData.Data.Orders) < 20 {
							break
						}
					}
				}

			}
		}
	}
}
