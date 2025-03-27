# dt-geo-converter

The **dt-geo-converter** CLI tool converts CSV files exported from spreadsheets into CWL workflows and relationship data. It features a modular command structure to simplify tasks like initializing the database, converting workflows, generating RO‑Crate templates.

## Features

- **Modular Commands:**  
  The tool is split into distinct subcommands for ease of use:

  - **init-db:** Initialize or update the database using your CSV files.
  - **convert:** Convert one or multiple workflows from the database into CWL descriptions and generate workflow graphs. A conversion-specific log file is saved alongside the generated files.
  - **generate-ro-crate:** Generate an RO‑Crate metadata package from a CWL file (work in progress).

- **CSV to CWL Conversion:**  
  Converts spreadsheet data into CWL workflows, generating separate CWL and DOT files for each step. A log file is created for each conversion, capturing all messages produced during that run.

- **Database Initialization and Update:**  
  Use the `init-db` command with the required `--dir` flag to initialize the database from CSV files. The directory can either contain the CSV files directly or hold subdirectories with the expected CSV files. The `--update` flag resets and reinitialize the database if needed.

- **RO‑Crate Template Generation (WIP):**  
  Generates an RO‑Crate metadata package from the CWL description (work in progress).

- **Automated README Generation:**  
  A README.md is automatically created in the workflow directory. This README includes an overview of the files, detected issues (if any), and a pointer to the conversion log file.

## Installation

1. **Download the Binary:**  
   Grab the binary for your platform from the [releases](https://github.com/Marco-Salvi/dt-geo-db/releases) section.

> [!NOTE]
> As of right now the tool has not yet been tested on Windows. If you encounter any problems, feel free to open an issue in the repository.

2. **Prepare Your CSV Files:**  
   Create a directory containing the CSV files exported from your spreadsheets. This directory should either contain the CSV files directly (e.g., `wf.csv`, `wf_wf.csv`, etc.) or have subdirectories where each contains the expected CSV files. See the [example](./data) in the repository for guidance.

## Usage

### Initializing the Database

Initialize (or update) your database with CSV data using:

```bash
dt-geo-converter init-db --dir <csv_directory> [--db <database_file>] [--update]
```

- **--dir:** Required. The directory containing your CSV files or subdirectories with CSV files.
- **--db:** Optional. Path to the database file (default: `./db.db`).
- **--update:** Optional. Reset and reinitialize the database if it already exists.

### Converting Workflows

Convert workflows from your database into CWL descriptions and workflow graphs. Each conversion run generates a log file (`log.log`) in the corresponding workflow's output directory:

```bash
dt-geo-converter convert --wf <workflow_id> [--db <database_file>] [--dir <csv_directory>] [--update]
```

Or to process all workflows:

```bash
dt-geo-converter convert --all [--db <database_file>] [--dir <csv_directory>] [--update]
```

- **--wf:** Specify the workflow ID to process (required if not using `--all`).
- **--all:** Process all workflows in the database.
- **--update:** Reset and reinitialize the database before conversion (requires the `--dir` flag).
- **Note:** The conversion-specific logs capture only the messages generated during that run.

For detailed options, run:

```bash
dt-geo-converter convert --help
```

## Reporting Issues

If you encounter any errors or issues while using the tool, please open an issue in the repository.
