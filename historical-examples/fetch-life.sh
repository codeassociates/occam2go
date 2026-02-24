#!/bin/bash
#
# Downloads the Life example from the "Programming in occam 2" book website
# and applies modifications needed to build with occam2go and the kroc course module.
#
# The original source is copyright Geraint Jones & Michael Goldsmith 1988, 2001.
# See: https://www.cs.ox.ac.uk/geraint.jones/publications/book/Pio2/code-life.txt
#
# Changes applied:
#   - Add copyright/attribution header and adaptation notes
#   - Add helper procedures (write.string, write.small.int) replacing book library
#   - Set board dimensions (array.width/array.height) to 20
#   - Replace clear.screen/move.cursor to avoid book library dependencies
#   - Add channel direction annotations (? and !) for occam2go compatibility
#   - Wrap main body in PROC life(...) declaration
#
# All line-number references are to the original downloaded file, which has not
# changed since 1988. Edits are applied bottom-to-top so that line numbers of
# earlier edits remain valid.

set -euo pipefail

URL="https://www.cs.ox.ac.uk/geraint.jones/publications/book/Pio2/code-life.txt"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT="${SCRIPT_DIR}/life.occ"

echo "Downloading original life.occ from book website..."
if ! curl -sfS -o "${OUTPUT}" "${URL}"; then
    echo "Error: failed to download ${URL}" >&2
    exit 1
fi

echo "Applying modifications for occam2go..."

# We apply edits bottom-to-top so earlier line numbers stay correct.

# --- Main body (lines 396-411): wrap in PROC life, indent, rename params ---
# Insert PROC header before line 396
sed -i '396i\PROC life (CHAN BYTE keyboard?, screen!, error!)' "${OUTPUT}"
# Indent lines 397-412 (the original 396-411, now shifted by 1) by 2 spaces
sed -i '397,412s/^/  /' "${OUTPUT}"
# Rename terminal.keyboard -> keyboard, terminal.screen -> screen in that range
sed -i '397,412s/terminal\.keyboard/keyboard/' "${OUTPUT}"
sed -i '397,412s/terminal\.screen/screen/' "${OUTPUT}"
# Append closing colon after line 412
sed -i '412a\:' "${OUTPUT}"
# Append trailing blank line
sed -i '413a\\' "${OUTPUT}"

# --- controller (line 345): add channel direction annotations ---
sed -i '345s/keyboard, screen,/keyboard?, screen!,/' "${OUTPUT}"

# --- editor (line 290): add channel direction annotations ---
sed -i '290s/keyboard, screen,/keyboard?, screen!,/' "${OUTPUT}"

# --- generation (line 208): add ! to screen param ---
sed -i '208s/screen,/screen!,/' "${OUTPUT}"

# --- display.activity (line 188): add ! to screen param ---
sed -i '188s/screen,/screen!,/' "${OUTPUT}"

# --- display.state (line 150): add ! to screen param ---
sed -i '150s/screen,/screen!,/' "${OUTPUT}"

# --- clean.up.display (line 146): add ! to screen param ---
sed -i '146s/screen)/screen!)/' "${OUTPUT}"

# --- initialize.display (line 141): add ! to screen param ---
sed -i '141s/screen)/screen!)/' "${OUTPUT}"

# --- move.cursor (lines 124-131): replace signature and body ---
# Line 124: add ! to terminal param
sed -i '124s/terminal,/terminal!,/' "${OUTPUT}"
# Lines 126-131: delete old body (DATA.ITEM/write.formatted), insert new
sed -i '126,131d' "${OUTPUT}"
# Insert new body at line 126 (after the comment on line 125)
sed -i '125a\
  -- outputs ANSI escape sequence: ESC [ row ; col H\
  SEQ\
    terminal ! BYTE #1B\
    terminal ! '"'"'['"'"'\
    write.small.int(terminal, y + 1)\
    terminal ! '"'"';'"'"'\
    write.small.int(terminal, x + 1)\
    terminal ! '"'"'H'"'"'' "${OUTPUT}"

# --- clear.screen (lines 119-121): replace implementation ---
# Line 119: add ! to terminal param
sed -i '119s/terminal)/terminal!)/' "${OUTPUT}"
# Line 120: replace comment (add explicit chars description)
sed -i '120s/-- clear screen sequence for an ANSI terminal/-- clear screen sequence for an ANSI terminal: ESC [ 2 J/' "${OUTPUT}"
# Line 121: delete old one-liner body
sed -i '121d' "${OUTPUT}"
# Insert new multi-line body after line 120
sed -i '120a\
  SEQ\
    terminal ! BYTE #1B\
    terminal ! '"'"'['"'"'\
    terminal ! '"'"'2'"'"'\
    terminal ! '"'"'J'"'"'' "${OUTPUT}"

# --- array dimensions (lines 8-9): replace ... with 20 ---
sed -i '8s/IS \.\.\. :/IS 20 :/' "${OUTPUT}"
sed -i '9s/IS \.\.\. :/IS 20 :/' "${OUTPUT}"

# --- Insert adaptation comment and helper procs after line 2 ---
sed -i '2a\
--  Adapted for occam2go: replaced book-library functions\
--  (write.string, write.formatted, DATA.ITEM) with inline\
--  definitions; added terminal.keyboard/terminal.screen declarations.\
--\
\
--\
--  helper procedures (replaces book standard library)\
--\
\
PROC write.string(CHAN OF BYTE out!, VAL []BYTE s)\
  SEQ i = 0 FOR SIZE s\
    out ! s[i]\
:\
\
PROC write.small.int(CHAN OF BYTE out!, VAL INT n)\
  -- outputs a small non-negative integer (0..999) as decimal digits\
  IF\
    n >= 100\
      SEQ\
        out ! BYTE ((n / 100) + (INT '"'"'0'"'"'))\
        out ! BYTE (((n / 10) \\ 10) + (INT '"'"'0'"'"'))\
        out ! BYTE ((n \\ 10) + (INT '"'"'0'"'"'))\
    n >= 10\
      SEQ\
        out ! BYTE ((n / 10) + (INT '"'"'0'"'"'))\
        out ! BYTE ((n \\ 10) + (INT '"'"'0'"'"'))\
    TRUE\
      out ! BYTE (n + (INT '"'"'0'"'"'))\
:' "${OUTPUT}"

# --- Insert copyright/attribution header at the very top ---
sed -i '1i\
--  Code copied from Programming in occam®2\
--  © Geraint Jones, Michael Goldsmith 1988, 2001.\
--  Permission is granted to copy this material for private study; for other uses please contact occam-book@comlab.ox.ac.uk\
--' "${OUTPUT}"

echo "Done. Output written to: ${OUTPUT}"
