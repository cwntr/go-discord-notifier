package fourchan

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cwntr/go-discord-notifier/pkg/common"
	"github.com/cwntr/go-discord-notifier/pkg/discord"
	"github.com/gocarina/gocsv"
	"github.com/grokify/html-strip-tags-go"
	"github.com/sirupsen/logrus"
)

const (
	ThreadsCSVFile = "threads.csv"
	CSVSeparator   = '|'

	maxLengthContent = 500
)

var (
	logger *logrus.Logger
	cfg    common.Config
)

type Thread struct {
	Id                 int    `csv:"id"`
	Timestamp          int    `csv:"timestamp"`
	Subject            string `csv:"subject"`
	Author             string `csv:"author"`
	Content            string `csv:"content"`
	Replies            int    `csv:"replies"`
	Images             int    `csv:"images"`
	LastReplyId        int    `csv:"last_reply_id"`
	LastReplyTimestamp int    `csv:"last_reply_time"`
	LastReplyAuthor    string `csv:"last_reply_author"`
	LastReplyContent   string `csv:"last_reply"`
}

func init() {
	logger = logrus.New()

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.LazyQuotes = true
		r.Comma = CSVSeparator
		return r
	})

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.UseCRLF = true
		writer.Comma = CSVSeparator
		return gocsv.NewSafeCSVWriter(writer)
	})
}

// Process func reads local csv file as a base for old files and requests the 4chan API for new entries. Completely new thread IDs will
// be treated as new found threads and for same thread IDs the last timestamp will be used to determine if an update (thread reply) occurred.
func Process(config common.Config) error {
	//before processing
	sanitizeCSVFile()

	cfg = config
	clientsFile, err := os.OpenFile(ThreadsCSVFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		logger.Errorf("unable to open csv file, err:%v", err)
		return err
	}
	defer clientsFile.Close()

	var oldThreads []Thread
	if err := gocsv.UnmarshalFile(clientsFile, &oldThreads); err != nil {
		logger.Errorf("unable to unmarshal csv file, err:%v", err)
	}

	if _, err := clientsFile.Seek(0, 0); err != nil {
		logger.Errorf("unable to reset csv file, err:%v", err)
		return err
	}

	currentThreads, err := GetLatestThreads()
	if err != nil {
		logger.Errorf("unable to reset csv file, err:%v", err)
		return err
	}

	processThreads(oldThreads, currentThreads)
	err = gocsv.MarshalFile(&currentThreads, clientsFile)
	if err != nil {
		logger.Errorf("unable to gocsv.MarshalFile, err:%v", err)
		return err
	}

	//after processing
	sanitizeCSVFile()
	return nil
}

// sanitizeCSVFile func will iterate through every line of the csv and removes all lines that do not match the csv format
func sanitizeCSVFile() {
	// sanitize
	clientsFile2, err := os.OpenFile(ThreadsCSVFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		logger.Errorf("unable os.OpenFile(ThreadsCSVFile, err:%v", err)
	}
	defer clientsFile2.Close()

	scanner := bufio.NewScanner(clientsFile2)
	cnt := 0
	var rows []string
	for scanner.Scan() {
		cnt++
		if cnt == 1 {
			rows = append(rows, scanner.Text())
			continue
		}
		if isCorrectCSVFormat(scanner.Text()) {
			rows = append(rows, scanner.Text())
		}
	}
	content := strings.Join(rows, "\n")
	err = ioutil.WriteFile(ThreadsCSVFile, []byte(content), os.ModePerm)
	if err != nil {
		logger.Errorf("unable to ioutil.WriteFile(ThreadsCSVFile, err:%v", err)
	}
}

// isCorrectCSVFormat func checks whether the all lines of the csv have the correct format considering the separator
func isCorrectCSVFormat(str string) bool {
	sepStr := strings.Replace(strconv.QuoteRune(CSVSeparator), "'", "", -1)
	var re = regexp.MustCompile(fmt.Sprintf(`^(\d+)(\%s)(\d+)(\%s)`, sepStr, sepStr))
	for _, _ = range re.FindAllString(str, -1) {
		return true
	}
	return false
}

func processThreads(old []Thread, new []Thread) {
	var newThreadsIds []int
	var updatedThreadsIds []int
	for _, n := range new {
		isThreadFound := false
		for _, o := range old {
			if o.Id == n.Id {
				if n.Timestamp > o.Timestamp {
					updatedThreadsIds = append(updatedThreadsIds, n.Id)
				}
				isThreadFound = true
			}
		}
		if !isThreadFound {
			newThreadsIds = append(newThreadsIds, n.Id)
		}
	}

	for _, newId := range newThreadsIds {
		for _, n := range new {
			if n.Id == newId {
				discord.NotifyNewThread(n.Id, BuildLink(n.Id, 0), n.Subject, n.Content)
				logger.Infof("NEW THREAD: [%s], [%s] [%s]", n.Subject, n.Author, n.Content)
			}
		}
	}

	for _, updatedId := range updatedThreadsIds {
		for _, n := range new {
			if n.Id == updatedId {
				discord.NotifyUpdateThread(n.Id, BuildLink(n.Id, n.LastReplyId), n.Subject, n.Replies, n.LastReplyAuthor, n.LastReplyContent)
				logger.Infof("UPDATED THREAD: [%s], [link: %s] [replies:%d] [images:%d] [lastReply: %s, %d, %s", n.Subject, BuildLink(n.Id, n.LastReplyId), n.Replies, n.Images, n.LastReplyAuthor, n.LastReplyTimestamp, n.LastReplyContent)
			}
		}
	}
}

func BuildLink(id int, replyId int) string {
	if replyId == 0 {
		return fmt.Sprintf("%s/%d", BIZThreadLinkURL, id)
	} else {
		return fmt.Sprintf("%s/%d#pc%d", BIZThreadLinkURL, id, replyId)
	}
}

func GetLatestThreads() ([]Thread, error) {
	cl := NewClient()
	res, err := cl.GetCatalog()
	if err != nil {
		logger.Errorf("err: %v", err)
		return []Thread{}, err

	}
	var threads []APIThread
	var addedIDs []int
	for _, r := range *res {
		for _, rt := range r.Threads {
			for _, keyWord := range GetKeywords() {
				if strings.Contains(strings.ToLower(rt.Sub), strings.ToLower(keyWord)) || strings.Contains(strings.ToLower(rt.Com), strings.ToLower(keyWord)) {
					isFound := false
					for _, id := range addedIDs {
						if id == rt.No {
							isFound = true
						}
					}
					if !isFound {
						threads = append(threads, rt)
						addedIDs = append(addedIDs, rt.No)
					}
				}
			}
		}
	}
	return APIThreadToCSVThreads(threads), nil
}

func APIThreadToCSVThreads(apiThreads []APIThread) []Thread {
	var threads []Thread
	for _, t := range apiThreads {
		thread := Thread{}
		thread.Id = t.No
		thread.Timestamp = t.LastModified
		thread.Subject = html.UnescapeString(t.Sub)
		thread.Author = t.Name

		threadContent := t.Com
		threadContent = html.UnescapeString(threadContent)
		threadContent = strings.Replace(threadContent, "<br>", " ", -1)
		threadContent = strip.StripTags(threadContent)
		//new line to space
		re := regexp.MustCompile(`\r?\n`)
		threadContent = re.ReplaceAllString(threadContent, " ")
		threadContent = stripRegex(threadContent)
		if len(threadContent) > maxLengthContent {
			thread.Content = threadContent[:maxLengthContent] + "..."
		} else {
			thread.Content = threadContent
		}
		thread.Replies = t.Replies
		thread.Images = t.Images
		if len(t.LastReplies) > 0 {
			threadContent := t.LastReplies[len(t.LastReplies)-1].Com
			threadContent = html.UnescapeString(threadContent)
			threadContent = strings.Replace(threadContent, "<br>", " ", -1)
			threadContent = strip.StripTags(threadContent)
			threadContent = re.ReplaceAllString(threadContent, " ")
			threadContent = stripRegex(threadContent)

			thread.LastReplyId = t.LastReplies[len(t.LastReplies)-1].No
			if len(threadContent) > maxLengthContent {
				thread.LastReplyContent = threadContent[:maxLengthContent] + "..."
			} else {
				thread.LastReplyContent = threadContent
			}
			thread.LastReplyTimestamp = t.LastReplies[len(t.LastReplies)-1].Time
			thread.LastReplyAuthor = t.LastReplies[len(t.LastReplies)-1].Name
		}
		threads = append(threads, thread)
	}
	return threads
}

func stripRegex(in string) string {
	reg, _ := regexp.Compile(`[^a-zA-Z0-9\-_/='.,:>#~?&%+<@() ]+`)
	return reg.ReplaceAllString(in, "")
}

// PeriodicCheck will request the 4chan API regularly for latest posts
func PeriodicCheck(cfg common.Config) {
	err := Process(cfg)
	if err != nil {
		logger.Errorf("4chan Process, err: %v", err)
	}
	//give it some time to process new threads [might only need time for first time init]
	time.Sleep(time.Minute * 1)

	ticker := time.NewTicker(cfg.Interval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				err := Process(cfg)
				if err != nil {
					logger.Errorf("4chan Process, err: %v", err)
				}
			}
		}
	}()
}

// GetKeywords func returns the mandatory keywords defined from config
func GetKeywords() []string {
	return cfg.Keywords
}
