package runtime

import (
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"fmt"
	"os"
)

// symbolizer maps runtime PCs to source locations via DWARF.
type symbolizer struct {
	slide uint64
	units []symbolUnit
}

// symbolUnit is one DWARF compilation unit and its line table.
type symbolUnit struct {
	ranges [][2]uint64
	reader *dwarf.LineReader
}

const markerName = "senbon_marker"

// newSymbolizer computes the ASLR slide from the embedded marker and loads the
// target's DWARF line tables. Symbolization is best-effort: a missing marker is
// an error, but missing DWARF simply yields no source locations.
func newSymbolizer(binary, dsym string, runtimeMarker uint64) (*symbolizer, error) {
	linkedMarker, err := markerAddress(binary)
	if err != nil {
		return nil, err
	}
	if linkedMarker == 0 {
		return nil, fmt.Errorf("symbol %s not found in %s", markerName, binary)
	}
	result := &symbolizer{slide: runtimeMarker - linkedMarker}
	if data, err := dwarfData(binary, dsym); err == nil {
		result.units, _ = collectUnits(data)
	}
	return result, nil
}

// markerAddress finds the linked address of the marker in the binary symtab.
func markerAddress(binary string) (uint64, error) {
	if file, err := elf.Open(binary); err == nil {
		defer func() { _ = file.Close() }()
		symbols, err := file.Symbols()
		if err == nil {
			if marker := elfMarker(symbols); marker != 0 {
				return marker, nil
			}
		}
	}
	file, err := macho.Open(binary)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()
	return machoMarker(file.Symtab), nil
}

// dwarfData opens DWARF from the dSYM when present, else the binary itself.
func dwarfData(binary, dsym string) (*dwarf.Data, error) {
	path := binary
	if dsym != "" {
		if _, err := os.Stat(dsym); err == nil {
			path = dsym
		}
	}
	if file, err := elf.Open(path); err == nil {
		defer func() { _ = file.Close() }()
		return file.DWARF()
	}
	file, err := macho.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return file.DWARF()
}

func elfMarker(symbols []elf.Symbol) uint64 {
	for _, symbol := range symbols {
		if symbol.Name == markerName || symbol.Name == "_"+markerName {
			return symbol.Value
		}
	}
	return 0
}

func machoMarker(symbols *macho.Symtab) uint64 {
	if symbols == nil {
		return 0
	}
	for _, symbol := range symbols.Syms {
		if symbol.Name == markerName || symbol.Name == "_"+markerName {
			return symbol.Value
		}
	}
	return 0
}

func collectUnits(data *dwarf.Data) ([]symbolUnit, error) {
	var units []symbolUnit
	reader := data.Reader()
	for {
		entry, err := reader.Next()
		if err != nil {
			return nil, err
		}
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagCompileUnit {
			reader.SkipChildren()
			continue
		}
		ranges, err := data.Ranges(entry)
		if err != nil || len(ranges) == 0 {
			continue
		}
		line, err := data.LineReader(entry)
		if err != nil || line == nil {
			continue
		}
		units = append(units, symbolUnit{ranges: ranges, reader: line})
	}
	return units, nil
}

// line returns the source location for a runtime PC.
func (s *symbolizer) line(pc uint64) (string, uint64, bool) {
	pc -= s.slide
	for _, unit := range s.units {
		if !inRanges(unit.ranges, pc) {
			continue
		}
		var entry dwarf.LineEntry
		if err := unit.reader.SeekPC(pc, &entry); err != nil {
			return "", 0, false
		}
		if entry.File == nil {
			return "", 0, false
		}
		return entry.File.Name, uint64(entry.Line), true
	}
	return "", 0, false
}

func inRanges(ranges [][2]uint64, pc uint64) bool {
	for _, r := range ranges {
		if pc >= r[0] && pc < r[1] {
			return true
		}
	}
	return false
}
