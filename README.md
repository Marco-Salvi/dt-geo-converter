# dt-geo-converter

The CWL Converter CLI tool allows you to convert CSV files from spreadsheets into relationship data and, as a work in progress, generate ro-crate templates from CWL descriptions.

## Installation

1. **Download the Binary:**  
   Grab the binary for your platform from the [releases](https://github.com/Marco-Salvi/dt-geo-db/releases) section.

2. **Prepare Your CSV Files:**  
   Create a directory containing the CSV files exported from your spreadsheets. Ensure you remove any content that isn’t part of the relationships. Refer to the examples in the repository for guidance: [example](./wp5).

## Usage

### Converting CSV Files

To convert your CSV files, run:

```bash
dt-geo-converter convert
```

For details on available options, use:

```bash
dt-geo-converter convert --help
```

### Generating a ro-crate Template (WIP)

A work-in-progress command is available to generate a ro-crate template from a CWL description. For usage details, run:

```bash
dt-geo-converter generate-ro-crate --help
```

## Reporting Issues

If you encounter any errors or issues while running the tool, please open an issue in the repository.
