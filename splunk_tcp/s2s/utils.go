package s2s

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	BUF = &bytes.Buffer{}
)

func encodeString(tosend string) []byte {
	buf := &bytes.Buffer{}
	l := uint32(len(tosend) + 1)
	binary.Write(buf, binary.BigEndian, l)
	binary.Write(buf, binary.BigEndian, []byte(tosend))
	binary.Write(buf, binary.BigEndian, []byte{0})
	return buf.Bytes()
}

func encodeKeyValue(key, value string) []byte {
	buf := &bytes.Buffer{}
	buf.Write(encodeString(key))
	buf.Write(encodeString(value))
	return buf.Bytes()
}

// EncodeEvent encodes a full Splunk event
func EncodeEvent(line map[string]string, buf *bytes.Buffer) *bytes.Buffer {
	// buf = &bytes.Buffer{}

	var msgSize uint32
	msgSize = 8 // Two unsigned 32 bit integers included, the number of maps and a 0 between the end of raw the _raw trailer
	maps := make([][]byte, 0)

	for k, v := range line {
		switch k {
		case "source":
			encodedSource := encodeKeyValue("MetaData:Source", "source::"+v)
			maps = append(maps, encodedSource)
			msgSize += uint32(len(encodedSource))
		case "sourcetype":
			encodedSourcetype := encodeKeyValue("MetaData:Sourcetype", "sourcetype::"+v)
			maps = append(maps, encodedSourcetype)
			msgSize += uint32(len(encodedSourcetype))
		case "host":
			encodedHost := encodeKeyValue("MetaData:Host", "host::"+v)
			maps = append(maps, encodedHost)
			msgSize += uint32(len(encodedHost))
		case "index":
			encodedIndex := encodeKeyValue("_MetaData:Index", v)
			maps = append(maps, encodedIndex)
			msgSize += uint32(len(encodedIndex))
		case "_raw":
			break
		default:
			encoded := encodeKeyValue(k, v)
			maps = append(maps, encoded)
			msgSize += uint32(len(encoded))
		}
	}

	encodedRaw := encodeKeyValue("_raw", line["_raw"])
	msgSize += uint32(len(encodedRaw))
	encodedRawTrailer := encodeString("_raw")
	msgSize += uint32(len(encodedRawTrailer))
	encodedDone := encodeKeyValue("_done", "_done")
	msgSize += uint32(len(encodedDone))

	binary.Write(buf, binary.BigEndian, msgSize)
	binary.Write(buf, binary.BigEndian, uint32(len(maps)+2)) // Include extra map for _done key and one for _raw
	for _, m := range maps {
		binary.Write(buf, binary.BigEndian, m)
	}
	binary.Write(buf, binary.BigEndian, encodedDone)
	binary.Write(buf, binary.BigEndian, encodedRaw)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, encodedRawTrailer)

	return buf
}

func InterfaceToString(i interface{}) (string, error) {
	switch v := i.(type) {
	case string:
		return v, nil
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v), nil
	case map[string]interface{}:
		jsonString, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonString), nil
	default:
		jsonString, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonString), nil
	}
}

func ReadYAMLFile(filename string, result interface{}) error {
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlData, result)
	if err != nil {
		return err
	}

	return nil
}
