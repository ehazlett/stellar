package guid

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type GUID [16]byte

func New() GUID {
	g := GUID{}
	_, err := io.ReadFull(rand.Reader, g[:])
	if err != nil {
		panic(err)
	}
	return g
}

func (g GUID) String() string {
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x-%02x", g[3], g[2], g[1], g[0], g[5], g[4], g[7], g[6], g[8:10], g[10:])
}

func FromString(s string) GUID {
	if strings.IndexAny(s, "\"") != 0 {
		s = "\"" + s
	}
	if strings.LastIndexAny(s, "\"") != len(s) - 1 {
		s = s + "\""
	}
	var resultGuid GUID
	err := json.Unmarshal([]byte(s), &resultGuid)
	if err != nil {
		panic(err)
	}
	return resultGuid
}

func (g *GUID) UnmarshalJSON(data []byte) error {
	dataString := string(data)
	dataString = strings.Trim(dataString, "\"")
	// Example: 14b90286-4191-4d7a-807b-434b78ba0496
	// First 8
	value, err := strconv.ParseInt(dataString[0:2], 16, 16)
	if err != nil {
		return err
	}
	g[3] = byte(value)
	value, err = strconv.ParseInt(dataString[2:4], 16, 16)
	if err != nil {
		return err
	}
	g[2] = byte(value)
	value, err = strconv.ParseInt(dataString[4:6], 16, 16)
	if err != nil {
		return err
	}
	g[1] = byte(value)
	value, err = strconv.ParseInt(dataString[6:8], 16, 16)
	if err != nil {
		return err
	}
	g[0] = byte(value)
	// Next 4
	value, err = strconv.ParseInt(dataString[9:11], 16, 16)
	if err != nil {
		return err
	}
	g[5] = byte(value)
	value, err = strconv.ParseInt(dataString[11:13], 16, 16)
	if err != nil {
		return err
	}
	g[4] = byte(value)
	// Next 4
	value, err = strconv.ParseInt(dataString[14:16], 16, 16)
	if err != nil {
		return err
	}
	g[7] = byte(value)
	value, err = strconv.ParseInt(dataString[16:18], 16, 16)
	if err != nil {
		return err
	}
	g[6] = byte(value)
	// Next 4
	value, err = strconv.ParseInt(dataString[19:21], 16, 16)
	if err != nil {
		return err
	}
	g[8] = byte(value)
	value, err = strconv.ParseInt(dataString[21:23], 16, 16)
	if err != nil {
		return err
	}
	g[9] = byte(value)
	// Last 12
	value, err = strconv.ParseInt(dataString[24:26], 16, 16)
	if err != nil {
		return err
	}
	g[10] = byte(value)
	value, err = strconv.ParseInt(dataString[26:28], 16, 16)
	if err != nil {
		return err
	}
	g[11] = byte(value)
	value, err = strconv.ParseInt(dataString[28:30], 16, 16)
	if err != nil {
		return err
	}
	g[12] = byte(value)
	value, err = strconv.ParseInt(dataString[30:32], 16, 16)
	if err != nil {
		return err
	}
	g[13] = byte(value)
	value, err = strconv.ParseInt(dataString[32:34], 16, 16)
	if err != nil {
		return err
	}
	g[14] = byte(value)
	value, err = strconv.ParseInt(dataString[34:36], 16, 16)
	if err != nil {
		return err
	}
	g[15] = byte(value)

	return nil
}