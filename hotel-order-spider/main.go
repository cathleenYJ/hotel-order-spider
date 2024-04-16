package main

import (
	"fmt"
	"gin/getOrders"
	"time"

	"github.com/spf13/viper"
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
}

func main() {

	viper.SetConfigFile("config_traiwan.yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	period := viper.GetString("period")

	// period轉為時間
	timeFormat := "2006-01"
	startTime, err := time.Parse(timeFormat, period)
	if err != nil {
		fmt.Println("Error parsing period:", err)
		return
	}

	// 設定 dateFrom
	dateFrom := startTime.Format("2006-01-02")

	// 設定 dateTo
	lastDayOfMonth := startTime.AddDate(0, 1, -1)
	dateTo := lastDayOfMonth.Format("2006-01-02")

	accounts := viper.Get("account").([]interface{})
	if accounts == nil {
		fmt.Println("無法取得 account")
		return
	}

	for _, acc := range accounts {
		account := acc.(map[string]interface{})

		accountName, ok := account["name"].(string)
		if !ok {
			fmt.Println("無法取得 account name")
			continue
		}

		fmt.Printf("accountName: %s\n", accountName)

		if platformRaw, ok := account["platform"]; ok {
			platforms, ok := platformRaw.([]interface{})
			if !ok || platforms == nil {
				fmt.Println("無法取得 platform")
				continue
			}

			for _, platformRaw := range platforms {
				platform, ok := platformRaw.(map[string]interface{})
				if !ok || platform == nil {
					fmt.Println("無法取得 platform")
					continue
				}

				platformName, ok := platform["name"].(string)
				if !ok {
					fmt.Println("無法取得 platform name")
					continue
				}

				fmt.Printf("platformName: %s\n", platformName)

				if platformName == "Booking" {
					getOrders.GetBooking(platform, platformName, period, dateFrom, dateTo)
				}

				if platformName == "Ctrip" {
					getOrders.GetCtrip(platform, platformName, accountName, dateFrom, dateTo)
				}

				if platformName == "Hostelworld" {
					getOrders.GetHostelworld(platform, platformName, dateFrom, dateTo)
				}

				if platformName == "NewSIM" {
					getOrders.GetNewSIM(platform, dateFrom, dateTo)
				}

				if platformName == "Owlting" {
					getOrders.GetOwlting(platform, dateFrom, dateTo)
				}

				if platformName == "Traiwan" {
					getOrders.GetTraiwan(platform, dateFrom, dateTo)
				}

			}
		}
	}
}
