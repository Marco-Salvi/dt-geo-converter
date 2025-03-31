package commands

import (
	"dt-geo-converter/logger"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Spreadsheet struct {
	Name   string
	Id     string
	Sheets []Sheet
}

// Sheet represents a single sheet with its name and gid.
type Sheet struct {
	Name string
	Gid  string
}

var spreadsheets = []Spreadsheet{
	{
		Name: "WP8",
		Id:   "1PBNwRbdwKxIC62_qGbhlpFFenqEQc-gjWqd2uZiHCv8",
		Sheets: []Sheet{
			{"wf", "0"},
			{"wf_wf", "592166767"},
			{"st_wf", "1799804703"},
			{"st_st", "6566471"},
			{"ss_st", "1214315610"},
			{"ss_ss", "736239675"},
			{"dt_st", "1422982849"},
			{"dt_ss", "1334178335"},
		},
	},
	{
		Name: "WP6",
		Id:   "1QUgmvmuuK13x_ElTCXBNHyvzHV44WCoaNstLiTLVb7Q",
		Sheets: []Sheet{
			{"wf", "0"},
			{"wf_wf", "592166767"},
			{"st_wf", "1799804703"},
			{"st_st", "6566471"},
			{"ss_st", "1214315610"},
			{"ss_ss", "736239675"},
			{"dt_st", "1422982849"},
			{"dt_ss", "1334178335"},
		},
	},
	{
		Name: "WP5",
		Id:   "1Dfj4GXIJNwvTT3LgKlRkUAISHiC6gsFKTf6mj3pitLU",
		Sheets: []Sheet{
			{"wf", "0"},
			{"wf_wf", "592166767"},
			{"st_wf", "1799804703"},
			{"st_st", "6566471"},
			{"ss_st", "1214315610"},
			{"ss_ss", "736239675"},
			{"dt_st", "1422982849"},
			{"dt_ss", "1334178335"},
		},
	},
	{
		Name: "WP7",
		Id:   "1lnSpLPO2XDrGRUizuIcIOwCTqEHmLBKFKFwHJIDe0sI",
		Sheets: []Sheet{
			{"wf", "0"},
			{"wf_wf", "592166767"},
			{"st_wf", "1799804703"},
			{"st_st", "6566471"},
			{"ss_st", "1214315610"},
			{"ss_ss", "736239675"},
			{"dt_st", "1422982849"},
			{"dt_ss", "1334178335"},
		},
	},
}

// Base URL pattern for CSV export including the gid parameter.
var baseURL = "https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s"

// For each sheet, download it and put it into a specific dir inside a temp dir
func DownloadRemoteSheets() (string, error) {
	// Create temp directory in data folder
	dir, err := os.MkdirTemp("", "sheets-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	for _, spreadsheet := range spreadsheets {
		logger.Info("Fetching", spreadsheet.Name, "from Google Drive")
		sheetDir := filepath.Join(dir, spreadsheet.Name)

		// Create a dir inside the temp dir for the sheets of this spreadsheet
		err := os.Mkdir(sheetDir, 0755)
		if err != nil {
			return dir, fmt.Errorf("failed to create directory for spreadsheet %s: %w", spreadsheet.Name, err)
		}

		for _, sheet := range spreadsheet.Sheets {
			// Construct the URL for this particular sheet
			url := fmt.Sprintf(baseURL, spreadsheet.Id, sheet.Gid)
			logger.Debug("Fetching CSV for", spreadsheet.Name+"-"+sheet.Name, "from", url)

			// Fetch the CSV data
			resp, err := http.Get(url)
			if err != nil {
				logger.Error("Error fetching CSV for", spreadsheet.Name+"-"+sheet.Name, ":", err)
				continue
			}

			// Ensure response body is closed after use
			defer resp.Body.Close()

			// Check if response is successful
			if resp.StatusCode != http.StatusOK {
				logger.Error("Error fetching CSV for", spreadsheet.Name+"-"+sheet.Name, ": status code ", resp.StatusCode)
				continue
			}

			// Create a new CSV file with .csv extension
			filePath := filepath.Join(sheetDir, sheet.Name+".csv")
			file, err := os.Create(filePath)
			if err != nil {
				logger.Error("Error creating file for", spreadsheet.Name+"-"+sheet.Name, ":", err)
				continue
			}

			// Copy the response body directly to the file
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				logger.Error("Error writing CSV data for", spreadsheet.Name+"-"+sheet.Name, ":", err)
				file.Close() // Close file before continuing
				continue
			}

			// Close file after writing
			err = file.Close()
			if err != nil {
				logger.Error("Error closing file for", spreadsheet.Name+"-"+sheet.Name, ":", err)
			}
		}
	}

	return dir, nil
}
