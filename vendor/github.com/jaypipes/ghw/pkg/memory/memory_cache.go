//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package memory

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw/pkg/unitutil"
)

type CacheType int

const (
	CACHE_TYPE_UNIFIED CacheType = iota
	CACHE_TYPE_INSTRUCTION
	CACHE_TYPE_DATA
)

var (
	memoryCacheTypeString = map[CacheType]string{
		CACHE_TYPE_UNIFIED:     "Unified",
		CACHE_TYPE_INSTRUCTION: "Instruction",
		CACHE_TYPE_DATA:        "Data",
	}

	// NOTE(fromani): the keys are all lowercase and do not match
	// the keys in the opposite table `memoryCacheTypeString`.
	// This is done because of the choice we made in
	// CacheType:MarshalJSON.
	// We use this table only in UnmarshalJSON, so it should be OK.
	stringMemoryCacheType = map[string]CacheType{
		"unified":     CACHE_TYPE_UNIFIED,
		"instruction": CACHE_TYPE_INSTRUCTION,
		"data":        CACHE_TYPE_DATA,
	}
)

func (a CacheType) String() string {
	return memoryCacheTypeString[a]
}

// NOTE(jaypipes): since serialized output is as "official" as we're going to
// get, let's lowercase the string output when serializing, in order to
// "normalize" the expected serialized output
func (a CacheType) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(strings.ToLower(a.String()))), nil
}

func (a *CacheType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	key := strings.ToLower(s)
	val, ok := stringMemoryCacheType[key]
	if !ok {
		return fmt.Errorf("unknown memory cache type: %q", key)
	}
	*a = val
	return nil
}

type SortByCacheLevelTypeFirstProcessor []*Cache

func (a SortByCacheLevelTypeFirstProcessor) Len() int      { return len(a) }
func (a SortByCacheLevelTypeFirstProcessor) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByCacheLevelTypeFirstProcessor) Less(i, j int) bool {
	if a[i].Level < a[j].Level {
		return true
	} else if a[i].Level == a[j].Level {
		if a[i].Type < a[j].Type {
			return true
		} else if a[i].Type == a[j].Type {
			// NOTE(jaypipes): len(LogicalProcessors) is always >0 and is always
			// sorted lowest LP ID to highest LP ID
			return a[i].LogicalProcessors[0] < a[j].LogicalProcessors[0]
		}
	}
	return false
}

type SortByLogicalProcessorId []uint32

func (a SortByLogicalProcessorId) Len() int           { return len(a) }
func (a SortByLogicalProcessorId) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByLogicalProcessorId) Less(i, j int) bool { return a[i] < a[j] }

type Cache struct {
	Level     uint8     `json:"level"`
	Type      CacheType `json:"type"`
	SizeBytes uint64    `json:"size_bytes"`
	// The set of logical processors (hardware threads) that have access to the
	// cache
	LogicalProcessors []uint32 `json:"logical_processors"`
}

func (c *Cache) String() string {
	sizeKb := c.SizeBytes / uint64(unitutil.KB)
	typeStr := ""
	if c.Type == CACHE_TYPE_INSTRUCTION {
		typeStr = "i"
	} else if c.Type == CACHE_TYPE_DATA {
		typeStr = "d"
	}
	cacheIDStr := fmt.Sprintf("L%d%s", c.Level, typeStr)
	processorMapStr := ""
	if c.LogicalProcessors != nil {
		lpStrings := make([]string, len(c.LogicalProcessors))
		for x, lpid := range c.LogicalProcessors {
			lpStrings[x] = strconv.Itoa(int(lpid))
		}
		processorMapStr = " shared with logical processors: " + strings.Join(lpStrings, ",")
	}
	return fmt.Sprintf(
		"%s cache (%d KB)%s",
		cacheIDStr,
		sizeKb,
		processorMapStr,
	)
}
