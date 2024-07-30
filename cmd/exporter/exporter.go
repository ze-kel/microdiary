package exporter

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/ze-kel/microdiary/cmd/db"
)

func ComposeTextFromMessages(messages []db.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var sb strings.Builder

	var lastDates MessageDates

	for ind, msg := range messages {
		mDates := getDates(msg.Date)

		firstMessage := ind == 0
		sameYear, sameDay, sameMonth := lastDates.year == mDates.year, lastDates.day == mDates.day, lastDates.month == mDates.month

		if firstMessage || !sameYear {
			if !firstMessage {
				sb.WriteString("\n\n")
			}
			sb.WriteString(fmt.Sprintf("## %d \n", mDates.year))
		}

		if firstMessage || !sameMonth {
			if !firstMessage && sameYear {
				sb.WriteString("\n\n")
			}
			sb.WriteString(fmt.Sprintf("### %s \n", mDates.month))
		}

		if firstMessage || !sameDay {
			sb.WriteString(fmt.Sprintf("##### %d \n", mDates.day))
		}

		if firstMessage || !sameDay || getDiffInMinutes(mDates.date, lastDates.date) > 10 {
			sb.WriteString(fmt.Sprintf("%s \n", mDates.timeFormatted))
		}

		sb.WriteString(msg.Message)
		sb.WriteString("\n")

		lastDates = mDates

	}

	return sb.String()

}

type MessageDates struct {
	day           int
	month         string
	year          int
	date          int64
	timeFormatted string
}

func getDiffInMinutes(t1, t2 int64) float64 {
	diff := time.Unix(t1, 0).Sub(time.Unix(t2, 0))
	return math.Abs(diff.Minutes())
}

func getDates(date int64) MessageDates {
	t1 := time.Unix(date, 0).In(getLocation())

	h, m, _ := t1.Clock()

	return MessageDates{
		day:           t1.Day(),
		month:         t1.Month().String(),
		year:          t1.Year(),
		date:          date,
		timeFormatted: fmt.Sprintf("%02d:%02d", h, m),
	}

}

func getLocation() *time.Location {
	tz, _ := os.LookupEnv("TIMEZONE")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("Error when parsing location from env (TIMEZONE)", tz)
		return time.Now().Location()
	}
	return loc
}
