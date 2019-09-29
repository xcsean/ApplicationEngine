package etc

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strconv"

	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

// QueryConfig query a config from global config
func QueryConfig(category, key string) (string, bool) {
	return gc.getValue(category, key)
}

// InConfig tell whether the key in category and has a sub-string like 'pattern'
func InConfig(category, key, pattern string) bool {
	return gc.contains(category, key, pattern)
}

// HaveAddress tell whether myself have the addr or not
func HaveAddress(addr string) bool {
	_, ok := selfAddrs[addr]
	return ok
}

// CanProvideService tell whether myself can provide the service
func CanProvideService(division string) (bool, error) {
	app, server, _, err := svc.ParseDivision(division)
	if err != nil {
		return false, err
	}
	ip, _, _, _, err := QueryNode(app, server, division)
	if err != nil {
		return false, err
	}
	ok := HaveAddress(ip)
	return ok, nil
}

// CompareInt64WithConfig compare two int64 value, if the key isn't exist or not a number, use defaultValue
func CompareInt64WithConfig(category, key string, givenValue, defaultValue int64, handler func(int64, int64) bool) bool {
	v, ok := gc.getValue(category, key)
	if !ok {
		return handler(givenValue, defaultValue)
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return handler(givenValue, defaultValue)
	}
	return handler(givenValue, i)
}

// GetInt64WithDefault get a int64 value by categroy and key, if not exist, use the defaultValue
func GetInt64WithDefault(category, key string, defaultValue int64) int64 {
	v, ok := gc.getValue(category, key)
	if !ok {
		return defaultValue
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

// ReadFromXMLFile read context from a xml file
func ReadFromXMLFile(fileName string) ([]byte, error) {
	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewReader(bytes.NewBuffer(fileData))
	header, err := buf.Peek(3)
	if err == nil && header[0] == 239 && header[1] == 187 && header[2] == 191 {
		// remove the BOM header
		return fileData[3:], nil
	}

	return fileData, nil
}
