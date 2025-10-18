#!/usr/bin/env bash
# Usage: ./clear_comments.sh /path/to/dir
set -e

DIR="${1:-.}"

if [ ! -d "$DIR" ]; then
  echo "Error: '$DIR' is not a directory"
  exit 1
fi

# Function that strips // comments outside of quotes
strip_comments() {
  awk '{
    line=$0
    in_str=0
    esc=0
    for(i=1;i<=length(line);i++){
      c=substr(line,i,1)
      if(c=="\"" && !esc){in_str=!in_str}
      if(!in_str && substr(line,i,2)=="//"){
        print substr(line,1,i-1)
        next
      }
      if(c=="\\" && !esc){esc=1}else{esc=0}
    }
    print line
  }' "$1"
}

export -f strip_comments

find "$DIR" -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.c" -o -name "*.cpp" \) \
  ! -path "*/.git/*" ! -path "*/node_modules/*" ! -path "*/vendor/*" | while read -r file; do
    tmp=$(mktemp)
    strip_comments "$file" > "$tmp"
    mv "$tmp" "$file"
    echo "Cleaned $file"
done
