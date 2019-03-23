package googlesheet

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/golang/glog"
	bdaybot "github.com/j-fuentes/bdaybot-slack/api"
	"google.golang.org/api/sheets/v4"
)

// BdayDateRegExp is a regexp for the date forma in the dday list spreadsheet
var BdayDateRegExp = regexp.MustCompile(`(?P<month>\d+)/(?P<day>\d+)`)

type Reader struct {
	srv     *sheets.Service
	sheetID string
}

func NewReader(client *http.Client, sheetID string) (*Reader, error) {
	srv, err := sheets.New(client)
	if err != nil {
		return nil, err
	}

	return &Reader{
		srv:     srv,
		sheetID: sheetID,
	}, nil
}

func (r *Reader) GetBdays() ([]*bdaybot.Bday, error) {
	fmt.Println(r.sheetID)
	// The spreadsheet is supposed to have a format like:
	//  | Name   | Date   |
	//  | rick   | 12/21  |
	//  | alice  | 03/14  |

	readRange := "A2:B"
	resp, err := r.srv.Spreadsheets.Values.Get(r.sheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	bdays := []*bdaybot.Bday{}
	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("No data found")
	} else {
		glog.Info("Name, BDay:")
		for idx, row := range resp.Values {
			name := fmt.Sprintf("%s", row[0])
			if name == "" {
				return nil, fmt.Errorf("Empty name in row %d", idx+2)
			}

			rawdate := fmt.Sprintf("%s", row[1])

			glog.Info("%s, %s\n", name, rawdate)

			date, err := parseDate(rawdate)
			if err != nil {
				return nil, fmt.Errorf("Cannot parse date in row %d: %+v", idx+2, err)
			}

			bdays = append(bdays, &bdaybot.Bday{
				UserID: name,
				Date:   date,
			})
		}
	}

	return bdays, nil
}

func parseDate(date string) (*bdaybot.Bday_Date, error) {
	match := BdayDateRegExp.FindStringSubmatch(date)
	if len(match) != 3 {
		return nil, fmt.Errorf("Error parsing date %q", date)
	}

	month, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, err
	}

	day, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, err
	}

	return &bdaybot.Bday_Date{
		Month: int32(month),
		Day:   int32(day),
	}, nil
}
