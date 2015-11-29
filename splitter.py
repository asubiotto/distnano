import sys, os, argparse

#parses and returns arguments
def getArgs():
    parser = argparse.ArgumentParser(
        description='Hosts distributed nanocubes by splitting the input .csv file with specified cols and number of clusters')
    parser.add_argument(
        '-n', '--numclusters', type=int, help='Number of clusters', required=True)
    parser.add_argument(
        '-f', '--filename', type=str, help='.csv file path', required=True)
    parser.add_argument(
        '-s', '--sep', type=str, help='.csv column separator', required=True)
    parser.add_argument(
        '-t', '--time', type=str, help='time col name', required=True)
    parser.add_argument(
        '-a', '--lat', type=str, help='latitude col name', required=True)
    parser.add_argument(
        '-o', '--lon', type=str, help='longitude col name', required=True)
    parser.add_argument(
        '-c', '--cat', type=str, help='category col name', required=True)
    # Array for all arguments passed to script
    args = parser.parse_args()
    # Return all arg values
    return args.numclusters, args.filename, args.sep, args.time, args.lat, args.lon, args.cat

#splitcsv takes a csv filename in the $NANOCUBE_SRC/data and splits it into 
#numclusters csv files, each with the same header line
def splitcsv(filename, numclusters):
	numLines = sum(1 for line in open(filename))
	linesPerFile = numLines / numclusters
	count = 0
	firstNum = 0
	fn = filename.split('.csv')[0]+'_split'
	for line in open(filename, 'r'):
		if (count == 0):
			firstLine = line
		else:
			fileNum = min((count-1)/linesPerFile + 1, numclusters)
			if (fileNum != firstNum):
				f = open(fn + str(fileNum) + '.csv', 'w+')
				f.write(firstLine)
				firstNum = fileNum
			f.write(line)
		count += 1
	print(count)

#returns a random open port
def get_open_port():
        import socket
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.bind(("",0))
        s.listen(1)
        port = s.getsockname()[1]
        s.close()
        return port


#hostdmp convers the split csv files to dmp files, hosts them on ports 29512 and above, and removes the initial split .csv files
def hostdmp(filename, numclusters):
	ports = []
	fn = filename.split('.csv')[0]+'_split'
	for i in range(numclusters):
		fileNum = i + 1
		port = get_open_port()
		ports.append(port)
		dmp_command = "nanocube-binning-csv --sep=',' --timecol='time' --latcol='Latitude' --loncol='Longitude' \
				--catcol='crime' " + fn + str(fileNum) + '.csv > '\
				 + fn + str(fileNum) + '.dmp'
		os.system(dmp_command)
		host_command = 'cat ' + fn + str(fileNum) + '.dmp | nanocube-leaf -q ' + str(port) + ' -f 10000 &'
		os.system(host_command)
		rm_command = 'rm ' + fn + str(fileNum) + '.csv'
		os.system(rm_command)


def main(argv):
	numclusters, filename, sep, timecol, latcol, loncol, catcol = getArgs()
	print(numclusters, filename, sep, timecol, latcol, loncol, catcol)
	splitcsv(filename, numclusters)
	hostdmp(filename, numclusters)


if __name__=='__main__':
	main(sys.argv[1:])