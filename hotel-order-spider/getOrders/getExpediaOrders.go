package getOrders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GetExpediaReservationsAPIRequestBody struct {
	Query string `json:"query"`
}

type ExpediaReservationItem struct {
	GuestName        string  `json:"guest_name"`
	CheckInDate      string  `json:"check_in_date"`
	CheckOutDate     string  `json:"check_out_date"`
	BookDate         string  `json:"book_date"`
	RoomType         string  `json:"room_type"`
	PaymentType      string  `json:"payment_type"`
	Price            float64 `json:"price"`
	ReservationId    string  `json:"reservation_id"`
	ConfirmationCode string  `json:"confirmation_code"`
	Status           string  `json:"status"`
	HotelId          string  `json:"hotel_id"`
	RoomNights       int64   `json:"room_nights"`
}

type GetExpediaReservationsResponseBody struct {
	Data struct {
		ReservationSearchV2 struct {
			ReservationItems []struct {
				ReservationItemID string `json:"reservationItemId"`
				ReservationInfo   struct {
					StartDate      string `json:"startDate"`
					EndDate        string `json:"endDate"`
					CreateDateTime string `json:"createDateTime"`
					Product        struct {
						UnitName string `json:"unitName"`
					} `json:"product"`
					ReservationAttributes struct {
						StayStatus string `json:"stayStatus"`
					} `json:"reservationAttributes"`
				} `json:"reservationInfo"`
				Customer struct {
					GuestName string `json:"guestName"`
				} `json:"customer"`
				ConfirmationInfo struct {
					ProductConfirmationCode string `json:"productConfirmationCode"`
				} `json:"confirmationInfo"`
				TotalAmounts struct {
					TotalAmountForPartners struct {
						Value float64 `json:"value"`
					} `json:"totalAmountForPartners"`
				} `json:"totalAmounts"`
				PaymentInfo struct {
					EvcCardDetailsExist any `json:"evcCardDetailsExist"`
				} `json:"paymentInfo"`
			} `json:"reservationItems"`
		} `json:"reservationSearchV2"`
	} `json:"data"`
}

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

		url := "https://api.expediapartnercentral.com/supply/experience/gateway/graphql"
		var reqBody GetExpediaReservationsAPIRequestBody
		reqBody.Query = fmt.Sprintf("query getReservationsBySearchCriteria {reservationSearchV2(input: {propertyId: %s, booked: true, bookingItemId: null, canceled: true, confirmationNumber: null, confirmed: true, startDate: \"%s\", endDate: \"%s\", dateType: \"checkOut\", evc: false, expediaCollect: true, timezoneOffset: \"+08:00\", firstName: null, hotelCollect: true, isSpecialRequest: false, isVIPBooking: false, lastName: null, reconciled: false, readyToReconcile: false, returnBookingItemIDsOnly: false, searchParam: null, unconfirmed: true searchForCancelWaiversOnly: false }) { reservationItems{ reservationItemId reservationInfo {reservationTpid propertyId startDate endDate createDateTime brandDisplayName newReservationItemId country reservationAttributes {businessModel bookingStatus fraudCancelled fraudReleased stayStatus eligibleForECNoShowAndCancel strongCustomerAuthentication} specialRequestDetails accessibilityRequestDetails product {productTypeId unitName bedTypeName propertyVipStatus} customerArrivalTime {arrival}readyToReconcile epsBooking } customer {id guestName phoneNumber email emailAlias country} loyaltyInfo {loyaltyStatus vipAmenities} confirmationInfo {productConfirmationCode} conversationsInfo {conversationsSupported id unreadMessageCount conversationStatus cpcePartnerId}totalAmounts {totalAmountForPartners {value currencyCode}totalCommissionAmount {value currencyCode}totalReservationAmount {value currencyCode}propertyBookingTotal {value currencyCode}totalReservationAmountInPartnerCurrency {value currencyCode}}reservationActions {requestToCancel {reason actionSupported actionUnsupportedBehavior {hide disable}}changeStayDates {reason actionSupported}requestRelocation {reason actionSupported}actionAttributes {highFence}reconciliationActions {markAsNoShow {reason actionSupported actionUnsupportedBehavior {hide disable openVa}virtualAgentParameters {intentName taxonomyId}}undoMarkNoShow {reason actionSupported actionUnsupportedBehavior {hide disable}}changeCancellationFee {reason actionSupported actionUnsupportedBehavior {hide disable}}resetCancellationFee {reason actionSupported actionUnsupportedBehavior {hide disable}}markAsCancellation {reason actionSupported actionUnsupportedBehavior {hide disable}}undoMarkAsCancellation {reason actionSupported actionUnsupportedBehavior {hide disable}}changeReservationAmountsOrDates {reason actionSupported actionUnsupportedBehavior {hide disable}}resetReservationAmountsOrDates {reason actionSupported actionUnsupportedBehavior {hide disable}}}}reconciliationInfo {reconciliationDateTime reconciliationType}paymentInfo {evcCardDetailsExist expediaVirtualCardResourceId creditCardDetails { viewable viewCountLimit viewCountLeft viewCount hideCvvFromDisplay valid prevalidateCardOptIn cardValidationViewable inViewingWindow validationInfo {validationStatus validationType validationDate validationBy hasGuestProvidedNewCC newCreditCardReceivedDate is24HoursFromLastValidation } }}billingInfo {invoiceNumber }cancellationInfo {cancelDateTime cancellationPolicy {priceCurrencyCode costCurrencyCode policyType cancellationPenalties {penaltyCost penaltyPrice penaltyPerStayFee penaltyTime penaltyInterval penaltyStartHour penaltyEndHour }nonrefundableDatesList}}compensationDetails {reservationWaiverType reservationFeeAmounts {propertyWaivedFeeLineItem {costCurrency costAmount }}} searchWaiverRequest {serviceRequestId type typeDetails state orderNumber partnerId createdDate srConversationId lastUpdatedDate notes {text author {firstName lastName }}}} numOfCancelWaivers}}", hotelId, dateFrom, dateTo)
		jsonReqBody, _ := json.Marshal(reqBody)

		var ordersData GetExpediaReservationsResponseBody
		// err := json.Unmarshal([]byte(result), &ordersData)
		// if err != nil {
		// 	fmt.Println("JSON解碼錯誤:", err)
		// 	return
		// }
		if err := DoRequestAndGetResponse_expedia("POST", url, bytes.NewBuffer(jsonReqBody), cookie, &ordersData); err != nil {
			fmt.Println("123err:", err)
			return
		}
		fmt.Println("ordersData", ordersData)

		var resultData []ReservationsDB

		for _, reservation := range ordersData.Data.ReservationSearchV2.ReservationItems {
			var data ReservationsDB
			data.GuestName = reservation.Customer.GuestName
			data.Platform = "Expedia"
			data.CheckInDate = reservation.ReservationInfo.StartDate
			data.CheckOutDate = reservation.ReservationInfo.EndDate
			data.BookDate = reservation.ReservationInfo.CreateDateTime
			data.RoomType = reservation.ReservationInfo.Product.UnitName
			data.Price = reservation.TotalAmounts.TotalAmountForPartners.Value
			data.Commission = 0
			data.BookingId = reservation.ReservationItemID

			originalStatus := reservation.ReservationInfo.ReservationAttributes.StayStatus
			if originalStatus == "postStay" {
				data.ReservationStatus = "已成立"
			} else if originalStatus == "cancelled" {
				data.ReservationStatus = "已取消"
				if data.Price != 0 {
					data.ReservationStatus = "Chargeable cancellation"
				}
			} else if originalStatus == "markedAsNoShow" {
				data.ReservationStatus = "已取消"
				if data.Price != 0 {
					data.ReservationStatus = "Chargeable no show"
				}
			} else {
				data.ReservationStatus = originalStatus
			}

			startDate, _ := time.Parse("2006-01-02", reservation.ReservationInfo.StartDate)
			endDate, _ := time.Parse("2006-01-02", reservation.ReservationInfo.EndDate)
			roomNights := 0
			for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
				roomNights += 1
			}
			data.RoomNights = int64(roomNights) - 1

			data.HotelId = hotelId

			resultData = append(resultData, data)
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

func DoRequestAndGetResponse_expedia(method string, url string, reqBody io.Reader, cookie string, resBody interface{}) error {
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
