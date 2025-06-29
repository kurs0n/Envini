package upload

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func UploadFile(path string) {
	file, exist := fileExist(path)
	if !exist {
		return
	} else {
		defer file.Close()

		// hashmap, err := parseEnvFile(file)
	}
}

func fileExist(path string) (*os.File, bool) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil, false
	}

	return file, true
}

func parseEnvFile(file *os.File) (map[string]string, error) {
	data := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		key := strings.Split(line, "=")[0]
		value := strings.Split(line, "=")[1]
		data[key] = value
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return data, nil
}
