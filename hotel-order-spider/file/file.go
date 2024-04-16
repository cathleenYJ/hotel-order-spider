package file

import (
	"fmt"
	"os"
)

func AppendToFile(fileName, content string) {
	createFile(fileName)

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0666)
	defer f.Close()

	if err != nil {
		fmt.Println(err.Error())
	} else {
		f.WriteString(content)
	}
}

func createFile(fileName string) {
	if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
		os.Create(fileName)
	}
}
