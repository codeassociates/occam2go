# Historical Examples

This directory contains example programs from historical occam publications, adapted to build and run with occam2go.

## Life (Conway's Game of Life)

From *Programming in occam 2* by Geraint Jones and Michael Goldsmith (1988).
Book website: https://www.cs.ox.ac.uk/geraint.jones/publications/book/Pio2/

The original source is copyrighted and not included in this repository. To obtain it, run the fetch script which downloads it from the book website and applies the modifications needed for occam2go:

```bash
./historical-examples/fetch-life.sh
```

This produces `historical-examples/life.occ`, which can then be transpiled and run:

```bash
./occam2go -o life.go historical-examples/life.occ
go run life.go
```

### Controls

- **E** — enter editor mode (use arrow keys to move, space/asterisk to toggle cells, Q to exit editor)
- **R** — run (free-running evolution)
- **S** — stop
- **Any other key** — single step
- **Q** — quit
