package exporter

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/ze-kel/microdiary/cmd/db"
)

type MessagesToFilesMap = map[int64][]db.FilesForMessages
type FileNameToFileNameMap = map[string]string

func buildFilesMap(files []db.FilesForMessages) (messagesToFiles MessagesToFilesMap, fileIdsToNames FileNameToFileNameMap) {
	messagesToFiles = make(MessagesToFilesMap)
	fileIdsToNames = make(FileNameToFileNameMap)

	sameDateFilesCounter := make(map[string]int)

	for _, file := range files {

		fileNameBase := formatFilenameDate(file.Date)
		fileNameFull := fileNameBase

		filesBeforeAtSameDate := sameDateFilesCounter[fileNameBase]

		if filesBeforeAtSameDate > 0 {
			fileNameFull = fmt.Sprintf("%s %d", fileNameBase, filesBeforeAtSameDate+1)
		}
		sameDateFilesCounter[fileNameBase] = sameDateFilesCounter[fileNameBase] + 1

		messagesToFiles[file.MessageId] = append(messagesToFiles[file.MessageId], file)
		fileIdsToNames[file.Filename] = fmt.Sprintf("%s%s", fileNameFull, file.Format)
	}

	return messagesToFiles, fileIdsToNames
}

type StorageOfFiles interface {
	ReadFile(ctx context.Context, fileName string) (file []byte, err error)
}

func CreateExportFile(ctx context.Context, messages []db.Message, files []db.FilesForMessages, storage StorageOfFiles) (fileData io.Reader, fileFormat string, err error) {

	filesMap, fileNamesMap := buildFilesMap(files)

	fullText := ComposeTextFromMessages(messages, filesMap, fileNamesMap)

	if len(files) == 0 {
		return bytes.NewReader([]byte(fullText)), ".md", nil
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	err = AddFileToZip(w, "export.md", []byte(fullText))
	if err != nil {
		return nil, "", err
	}

	for _, file := range files {
		bts, err := storage.ReadFile(ctx, file.Filename)
		if err != nil {
			return nil, "", err
		}
		err = AddFileToZip(w, fileNamesMap[file.Filename], bts)
		if err != nil {
			return nil, "", err
		}
	}

	err = w.Close()
	if err != nil {
		return nil, "", err
	}
	return bufio.NewReader(buf), ".zip", nil

}

func AddFileToZip(z *zip.Writer, filename string, bts []byte) error {
	f, err := z.Create(filename)
	if err != nil {
		return err
	}

	_, err = f.Write(bts)
	if err != nil {
		return err
	}

	return nil
}

func ComposeTextFromMessages(messages []db.Message, filesAttached MessagesToFilesMap, fileNames FileNameToFileNameMap) string {
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
			sb.WriteString(fmt.Sprintf("##### %d %s\n", mDates.day, mDates.weekday))
		}

		if firstMessage || !sameDay || getDiffInMinutes(mDates.date, lastDates.date) > 10 {
			sb.WriteString(fmt.Sprintf("%s \n", mDates.timeFormatted))
		}

		sb.WriteString(msg.Message)

		for _, file := range filesAttached[msg.MessageId] {
			n := fileNames[file.Filename]
			sb.WriteString(fmt.Sprintf("![[%s]]", n))
		}

		sb.WriteString("\n")

		lastDates = mDates

	}

	return sb.String()

}

func formatFilenameDate(date int64) string {
	d := getDates(date)

	return fmt.Sprintf("%d.%02d.%02d — %s", d.year, d.monthNumber, d.day, d.timeFormattedFilename)
}

func getDiffInMinutes(t1, t2 int64) float64 {
	diff := time.Unix(t1, 0).Sub(time.Unix(t2, 0))
	return math.Abs(diff.Minutes())
}

type MessageDates struct {
	day                   int
	weekday               string
	month                 string
	monthNumber           int
	year                  int
	date                  int64
	timeFormatted         string
	timeFormattedFilename string
}

func getDates(date int64) MessageDates {
	t1 := time.Unix(date, 0).In(getLocation())

	h, m, _ := t1.Clock()

	return MessageDates{
		day:                   t1.Day(),
		month:                 t1.Month().String(),
		monthNumber:           int(t1.Month()),
		year:                  t1.Year(),
		date:                  date,
		weekday:               t1.Weekday().String(),
		timeFormatted:         fmt.Sprintf("%02d:%02d", h, m),
		timeFormattedFilename: fmt.Sprintf("%02d-%02d", h, m),
	}

}

func getLocation() *time.Location {
	tz, _ := os.LookupEnv("TIMEZONE")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("Error when parsing Timezone from env %s", tz)
		return time.Now().Location()
	}
	return loc
}
