package main

import (
	"bufio"
	"flag"
	"log"
	"os"
)

func main() {
	var datafileName string
	var timeCol string
	var latCol string
	var lonCol string

	// TODO(asubiotto): Just one category column for now but change this.
	var catCol string
	var distributionDegree int

	flag.StringVar(
		&datafileName,
		"data",
		"",
		"Path to csv file containing data",
	)

	flag.StringVar(
		&timeCol,
		"timecol",
		"",
		"Name of column in csv file representing time",
	)

	flag.StringVar(
		&latCol,
		"latcol",
		"",
		"Name of column in csv file representing latitude",
	)

	flag.StringVar(
		&lonCol,
		"loncol",
		"",
		"Name of column in csv file representing longitude",
	)

	flag.StringVar(
		&catCol,
		"loncol",
		"",
		"Name of column in csv file representing category",
	)

	flag.
		flag.IntVar(
		&distributionDegree,
		"dd",
		1,
		"How many nanocube slave nodes to spawn",
	)

	flags.Parse()

	// Split the csv file into dd partitions.
	file, err := os.Open(datafileName)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	scanner := bufio.NewScanner(bufio.NewReader(file))

	// TODO(asubiotto): This is currently hard-coded for the crime dataset. Add
	// some code to count the number of lines.
	numLines := 50000

	// The name of the current partition file that we are working with.
	var curFile os.File

	// Read the first description line.
	for i := 0; scanner.Scan(); i++ {
		// TODO(asubiotto): This is going to work but could result in spawning
		// an extra server. Find a better way to do this.
		if i%(numLines/distributionDegree) == 0 {
			// Create new file with ext from datafileName.
			// I'm wasting time right now, get to the point.
			os.Create()
		}
	}
}
