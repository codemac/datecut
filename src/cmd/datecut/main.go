package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	var (
		datefield int
		datefmt   string
		timefield int
		timefmt   string
		fieldsep  string
		help      bool
		points    string
		filename  string
	)

	flag.IntVar(&datefield, "d", 1, "which field of the log line is the date")
	flag.IntVar(&timefield, "t", 2, "which field of the log line is the time")
	flag.StringVar(&datefmt, "df", "2006-01-02", "how is the date formatted")
	flag.StringVar(&timefmt, "tf", "15:04:05.999999", "how is the time formatted, go-format-style (fuck)")
	flag.StringVar(&fieldsep, "s", " ", "how are the fields separated")
	flag.StringVar(&points, "p", "", "dates to split on, comma separated, RFC3339 formatted")
	flag.StringVar(&filename, "f", "datecut.out", "filename to output to, with dates postfixed")
	flag.BoolVar(&help, "h", false, "print this dialog")

	flag.Parse()

	if help || points == "" {
		fmt.Printf("help = %s", help)
		fmt.Printf("points = %s", points)
		flag.Usage()
		os.Exit(0)
	}

	pls := strings.Split(points, ",")
	pl := make([]time.Time, len(pls)+1)

	for i := range pls {
		var err error
		pl[i], err = time.Parse(time.RFC3339, pls[i])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		// do everything in UTC
		pl[i] = pl[i].UTC()
	}

	// create "last" date that everything will be before
	pl[len(pl)-1] = time.Date(9999, 9, 9, 9, 9, 9, 9, time.UTC)

	// index in the pl array we're working on
	pli := 0

	file, err := datefile(filename, pl[pli])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		strs := strings.Split(line, fieldsep)
		d, err := time.Parse(datefmt, strs[datefield-1])
		if err != nil {
			fmt.Fprintln(file, line)
			continue
		}
		d = d.UTC()

		t, err := time.Parse(timefmt, strs[timefield-1])
		if err != nil {
			fmt.Fprintln(file, line)
			continue
		}
		t = t.UTC()

		// combine date and time! note that I don't use a separator here
		// so we don't run into stupid fucking bugs.

		c := time.Date(d.Year(), d.Month(), d.Day(),
			t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)

		if !c.Before(pl[pli]) {
			// split, print, and increment
			file.Close()
			pli++
			file, err = datefile(filename, pl[pli])
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}

		fmt.Fprintln(file, line)
	}
}

func datefile(filename string, t time.Time) (*os.File, error) {
	rfc3339_with_z := "2006-01-02T15:04:05Z"
	f := filename + "_" + t.UTC().Format(rfc3339_with_z)
	return os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
}
