package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func WriteYAMLFile(filename string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, yamlData, 0644)
	if err != nil {
		return err
	}

	return nil
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
	fmt.Println(result)
	return nil
}

func CheckMethod(r *http.Request, method string) bool {
	if r.Method != method {
		return false
	}
	return true
}

func CheckBody(r *http.Request, dataType interface{}) bool {
	err := json.NewDecoder(r.Body).Decode(dataType)
	if err != nil {
		return false
	}
	return true
}

func ParseDurationString(durationStr string) (time.Duration, error) {
	// Sử dụng regex để phân tích chuỗi thời gian
	re := regexp.MustCompile(`(\d+)([dhm])`)
	matches := re.FindAllStringSubmatch(durationStr, -1)

	// Thêm thời gian tương ứng với mỗi phần từ
	duration := time.Duration(0)
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}

		unit := match[2]
		switch unit {
		case "d":
			duration += time.Duration(value) * 24 * time.Hour
		case "h":
			duration += time.Duration(value) * time.Hour
		case "m":
			duration += time.Duration(value) * time.Minute
		}
	}

	return duration, nil
}

func TimeStringToCron(timeString string) (string, error) {
	// Parse giờ và phút từ chuỗi
	parts := strings.Split(timeString, "h")
	if len(parts) != 2 {
		return "", fmt.Errorf("Chuỗi không hợp lệ")
	}

	hour := parts[0]
	minute := parts[1]

	// Tạo biểu diễn crontab từ giờ và phút
	crontabString := fmt.Sprintf("0 %s %s * * *", minute, hour)

	return crontabString, nil
}

func IsSlice(data interface{}) ([]interface{}, bool) {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Slice {
		slice := make([]interface{}, value.Len())
		for i := 0; i < value.Len(); i++ {
			slice[i] = value.Index(i).Interface()
		}
		return slice, true
	}
	return nil, false
}
