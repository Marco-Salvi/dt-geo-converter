# dt-geo-converter

The **dt-geo-converter** CLI tool converts CSV files exported from spreadsheets into CWL workflows and relationship data.

## Installation

1. **Download the Binary:**  
   Grab the binary for your platform from the [releases](https://github.com/Marco-Salvi/dt-geo-db/releases) section.

> [!NOTE]
> As of right now the tool has not yet been tested on Windows. If you encounter any problems, feel free to open an issue in the repository.

2. **Prepare Your Data Source:**  
   You can use the tool with either:
   - Local CSV files: Create a directory containing the CSV files exported from your spreadsheets
   - Remote data: Import data directly from the Google Drive spreadsheets

## Usage

### Getting Help

For a complete list of available commands and their options:

```bash
dt-geo-converter --help
```

For help with a specific command:

```bash
dt-geo-converter [command] --help
```

### Checking the Version

To see which version of the tool you have installed, run:

```bash
dt-geo-converter version
```

### Development

During development you can use the provided `makefile` to run common tasks:

```bash
make build VERSION=dev
make vet
...
```

## Data Sources

The tool supports two primary data sources:

1. **Local CSV Files:**  
   You can use CSV files exported from spreadsheets. The directory should either contain the CSV files directly (e.g., wf.csv, wf_wf.csv, etc.) or have subdirectories where each contains the expected CSV files. See the [example](./data) in the repository for guidance.

2. **Google Drive Integration:**  
   The tool can directly access and download spreadsheet data from the DT-GEO Google Drive, eliminating the need for manual exports. You can update the online spreadsheets and re-initialize the local database of the tool to fetch your changes.

## Output

The tool generates various outputs including:

- CWL workflow descriptions
- Workflow graphs
- Conversion log files
- RO‑Crate metadata packages (Work In Progress)

## Reporting Issues

If you encounter any errors or issues while using the tool, please open an issue in the repository.
