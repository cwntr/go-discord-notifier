package fourchan

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func Test_Client(t *testing.T) {
	cl := NewClient()
	res, err := cl.GetCatalog()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	for _, r := range *res {
		for _, rt := range r.Threads {
			if strings.Contains(strings.ToLower(rt.Sub), "xsn") || strings.Contains(strings.ToLower(rt.Com), "xsn") {
				fmt.Printf("Thread: [%d] [%s] [replies:%d] [images:%d]\n", rt.No, rt.Sub, rt.Replies, rt.OmittedImages)
			}
		}
	}
}

func Test_FileSanitize(t *testing.T) {
	// sanitize
	clientsFile2, err := os.OpenFile(ThreadsCSVFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		logger.Errorf("unable to open csv file, err:%v", err)
	}
	defer clientsFile2.Close()

	scanner := bufio.NewScanner(clientsFile2)
	cnt := 0
	var rows []string
	for scanner.Scan() {
		cnt++
		if cnt == 1 {
			continue
		}
		fmt.Println(scanner.Text())
		if isCorrectCSVFormat(scanner.Text()) {
			rows = append(rows, scanner.Text())
		} else {
			fmt.Println("not match")
		}
	}
	content := strings.Join(rows, "\n")
	err = ioutil.WriteFile(ThreadsCSVFile, []byte(content), os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
}
