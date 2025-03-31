package commands

import (
	"dt-geo-converter/logger"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func GetAvailableWPs() []string {
	result := make([]string, 0)
	for _, sheet := range spreadsheets {
		result = append(result, sheet.Name)
	}
	return result
}

// Base URL pattern for CSV export including the gid parameter.
var baseURL = "https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s"

// DownloadRemoteSheets downloads the sheets for only the specified work packages.
// If "all" is passed in toLoad (case-insensitive), then all spreadsheets are downloaded.
func DownloadRemoteSheets(toLoad []string) (string, error) {
	// Create a temporary directory for the sheets.
	dir, err := os.MkdirTemp("", "sheets-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Determine if "all" is specified.
	loadAll := false
	for _, s := range toLoad {
		if strings.EqualFold(s, "all") {
			loadAll = true
			break
		}
	}

	// Loop over the spreadsheets and process only the specified ones.
	for _, spreadsheet := range spreadsheets {
		if !loadAll {
			found := slices.Contains(toLoad, spreadsheet.Name)
			if !found {
				logger.Debug("Skipping spreadsheet", spreadsheet.Name, "as it's not specified in toLoad")
				continue
			}
		}

		logger.Info("Fetching", spreadsheet.Name, "from Google Drive")
		sheetDir := filepath.Join(dir, spreadsheet.Name)
		if err := os.Mkdir(sheetDir, 0755); err != nil {
			return dir, fmt.Errorf("failed to create directory for spreadsheet %s: %w", spreadsheet.Name, err)
		}

		for _, sheet := range spreadsheet.Sheets {
			url := fmt.Sprintf(baseURL, spreadsheet.Id, sheet.Gid)
			logger.Debug("Fetching CSV for", spreadsheet.Name+"-"+sheet.Name, "from", url)

			resp, err := http.Get(url)
			if err != nil {
				logger.Error("Error fetching CSV for", spreadsheet.Name+"-"+sheet.Name, ":", err)
				continue
			}
			// Ensure the response body is closed explicitly after processing.
			if resp.StatusCode != http.StatusOK {
				logger.Error("Error fetching CSV for", spreadsheet.Name+"-"+sheet.Name, ": status code", resp.StatusCode)
				resp.Body.Close()
				continue
			}

			filePath := filepath.Join(sheetDir, sheet.Name+".csv")
			file, err := os.Create(filePath)
			if err != nil {
				logger.Error("Error creating file for", spreadsheet.Name+"-"+sheet.Name, ":", err)
				resp.Body.Close()
				continue
			}

			// Copy the CSV data to the file.
			_, err = io.Copy(file, resp.Body)
			file.Close()
			resp.Body.Close()
			if err != nil {
				logger.Error("Error writing CSV data for", spreadsheet.Name+"-"+sheet.Name, ":", err)
				continue
			}
		}
	}

	return dir, nil
}
