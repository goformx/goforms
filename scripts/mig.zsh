#!/bin/zsh

# Define key computing dates in Laravel timestamp format
key_dates=(
  "1970_01_01_000000"  # Unix epoch
  "1983_01_01_000001"  # TCP/IP standard adoption
  "1991_08_06_000002"  # WWW announcement by Tim Berners-Lee
  "2004_02_04_000003"  # Facebook launch
  "2007_06_29_000004"  # First iPhone release
  "2015_06_01_000005"  # Docker 1.0 release
  "2020_12_01_000006"  # GPT-3 impact
  "2025_05_15_000007"  # Today's date for completeness
)

# Move to migrations directory
cd migrations || exit 1

# Get migration files
migration_files=( *_*.sql )
num_files=${#migration_files[@]}

if (( num_files == 0 )); then
  echo "No migration files found!"
  exit 1
fi

# Rename each migration file with a historical timestamp
for i in {1..$num_files}; do
  base_name="${migration_files[$i]}"
  new_name="${key_dates[$i]}_${base_name#*_}"
  mv -- "$base_name" "$new_name"
  echo "Renamed: $base_name -> $new_name"
done

echo "Migration file renaming complete!"
