#!/bin/bash

# FX Backtesting Data Converter
# Converts USDJPY CSV files from format:
# date,time,open,high,low,close,volume
# To format:
# timestamp,open,high,low,close,volume

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to convert a single file
convert_file() {
    local input_file="$1"
    local output_file="$2"
    
    print_status "Converting $input_file -> $output_file"
    
    # Check if input file exists
    if [[ ! -f "$input_file" ]]; then
        print_error "Input file not found: $input_file"
        return 1
    fi
    
    # Create output directory if it doesn't exist
    local output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"
    
    # Convert the file using awk
    awk -F',' '{
        # Skip empty lines
        if (NF < 6) next
        
        # Convert date format from YYYY.MM.DD to YYYY-MM-DD
        gsub(/\./, "-", $1)
        
        # Combine date and time into timestamp
        timestamp = $1 " " $2 ":00"
        
        # Output in required format: timestamp,open,high,low,close,volume
        printf "%s,%s,%s,%s,%s,%s\n", timestamp, $3, $4, $5, $6, $7
    }' "$input_file" > "$output_file"
    
    # Check if conversion was successful
    if [[ $? -eq 0 ]]; then
        local line_count=$(wc -l < "$output_file")
        print_success "Converted $line_count lines to $output_file"
        return 0
    else
        print_error "Failed to convert $input_file"
        return 1
    fi
}

# Function to show file preview
preview_file() {
    local file="$1"
    local lines="${2:-5}"
    
    echo -e "\n${BLUE}Preview of $file (first $lines lines):${NC}"
    echo "----------------------------------------"
    head -n "$lines" "$file" 2>/dev/null || echo "File not found or empty"
    echo "----------------------------------------"
}

# Function to validate converted file
validate_file() {
    local file="$1"
    
    print_status "Validating $file"
    
    # Check if file exists and is not empty
    if [[ ! -s "$file" ]]; then
        print_error "Output file is empty or doesn't exist: $file"
        return 1
    fi
    
    # Check format of first few lines
    local invalid_lines=$(head -n 10 "$file" | awk -F',' '
        {
            # Check number of fields
            if (NF != 6) {
                print "Line " NR ": Wrong number of fields (" NF " instead of 6)"
                exit 1
            }
            
            # Check timestamp format (basic validation)
            if ($1 !~ /^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}$/) {
                print "Line " NR ": Invalid timestamp format: " $1
                exit 1
            }
            
            # Check that OHLC values are numeric
            for (i = 2; i <= 5; i++) {
                if ($i !~ /^[0-9]+\.?[0-9]*$/) {
                    print "Line " NR ": Invalid numeric value in field " i ": " $i
                    exit 1
                }
            }
        }
        END { if (NR > 0) print "OK" }
    ')
    
    if [[ "$invalid_lines" == "OK" ]]; then
        print_success "File validation passed"
        return 0
    else
        print_error "File validation failed: $invalid_lines"
        return 1
    fi
}

# Main conversion function
main() {
    local input_dir="${1:-testdata}"
    local output_dir="${2:-testdata/converted}"
    
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}    FX Backtesting Data Converter${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
    echo "Input directory:  $input_dir"
    echo "Output directory: $output_dir"
    echo ""
    
    # Find all USDJPY CSV files
    local files=($(find "$input_dir" -name "USDJPY_*.csv" -type f | sort))
    
    if [[ ${#files[@]} -eq 0 ]]; then
        print_warning "No USDJPY_*.csv files found in $input_dir"
        echo ""
        echo "Expected file pattern: USDJPY_YYYY_MM.csv"
        echo "Example: USDJPY_2024_01.csv"
        return 1
    fi
    
    print_status "Found ${#files[@]} files to convert"
    echo ""
    
    # Convert each file
    local converted=0
    local failed=0
    
    for input_file in "${files[@]}"; do
        local filename=$(basename "$input_file")
        local output_file="$output_dir/$filename"
        
        # Show preview of original file
        preview_file "$input_file" 3
        
        # Convert the file
        if convert_file "$input_file" "$output_file"; then
            # Validate the converted file
            if validate_file "$output_file"; then
                # Show preview of converted file
                preview_file "$output_file" 3
                ((converted++))
            else
                ((failed++))
            fi
        else
            ((failed++))
        fi
        
        echo ""
    done
    
    # Summary
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}              SUMMARY${NC}"
    echo -e "${BLUE}============================================${NC}"
    print_success "Successfully converted: $converted files"
    if [[ $failed -gt 0 ]]; then
        print_error "Failed to convert: $failed files"
    fi
    echo "Output directory: $output_dir"
    echo ""
    
    # Show total records
    if [[ $converted -gt 0 ]]; then
        local total_records=$(find "$output_dir" -name "USDJPY_*.csv" -exec wc -l {} + | tail -n 1 | awk '{print $1}')
        print_status "Total records converted: $total_records"
        
        # Suggest next steps
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "1. Test with backtester:"
        echo "   go run cmd/backtester/main.go --data $output_dir/USDJPY_2024_01.csv"
        echo ""
        echo "2. Use in your code:"
        echo "   candles, err := dataProvider.LoadCSVData(\"$output_dir/USDJPY_2024_01.csv\")"
    fi
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "FX Backtesting Data Converter"
        echo ""
        echo "Usage: $0 [input_dir] [output_dir]"
        echo ""
        echo "Arguments:"
        echo "  input_dir   Input directory containing USDJPY_*.csv files (default: testdata)"
        echo "  output_dir  Output directory for converted files (default: testdata/converted)"
        echo ""
        echo "Example:"
        echo "  $0 testdata testdata/converted"
        echo ""
        echo "File format conversion:"
        echo "  From: 2024.01.02,00:00,140.833,140.838,140.827,140.827,13"
        echo "  To:   2024-01-02 00:00:00,140.833,140.838,140.827,140.827,13"
        exit 0
        ;;
    --version|-v)
        echo "FX Backtesting Data Converter v1.0"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac