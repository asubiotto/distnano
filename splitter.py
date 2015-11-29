import sys, os, argparse, concurrent.futures, multiprocessing
from datetime import datetime

ports = []


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
def splitcsv(filename, numclusters, sep, timecol):
	numLines = sum(1 for line in open(filename))
	linesPerFile = numLines / numclusters
	count = 0
	firstNum = 0
	mindt = datetime.max
	maxdt = datetime.min
	fn = filename.split('.csv')[0]+'_split'
	for line in open(filename, 'r'):
		if (count == 0):
			firstLine = line
			minLine = line
			maxLine = line
			timecol_index = firstLine.split(sep).index(timecol)
		else:
			t = line.split(sep)[timecol_index]
			dt = datetime.strptime(t, "%m/%d/%Y %I:%M:%S %p")
			if (dt < mindt):
				mindt = dt
				minLine = line
			if (dt > maxdt):
				maxdt = dt
				maxLine = line
			fileNum = min((count-1)/linesPerFile + 1, numclusters)
			if (fileNum != firstNum):
				f = open(fn + str(fileNum) + '.csv', 'w+')
				f.write(firstLine)
				firstNum = fileNum
			f.write(line)
		count += 1
	for i in range(fileNum):
		f = open(fn + str(i + 1) + '.csv', 'a')
		f.write(minLine)
		f.write(maxLine)

#returns a random open port
def get_open_port():
        import socket
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.bind(("",0))
        s.listen(1)
        port = s.getsockname()[1]
        s.close()
        return port

#hostdmp convers the split csv files to dmp files, hosts the nanocube, and deletes the original split .csv file
def hostdmp(filename, sep, timecol, latcol, loncol, catcol, i):
	try:
		fileNum = i + 1
		port = get_open_port()
		ports.append(port)
		fn = filename.split('.csv')[0]+'_split'
		dmp_command = "nanocube-binning-csv --sep='"+sep+"' --timecol='"+timecol+"' --latcol='"+latcol+"' --loncol='"+loncol+"' \
				--catcol='"+catcol+"' " + fn + str(fileNum) + '.csv > '\
				 + fn + str(fileNum) + '.dmp'
		os.system(dmp_command)
		host_command = 'cat ' + fn + str(fileNum) + '.dmp | nanocube-leaf -q ' + str(port) + ' -f 10000 &'
		os.system(host_command)
		#rm_command = 'rm ' + fn + str(fileNum) + '.csv'
		#os.system(rm_command)
	except:
		print('error with port')
	


def main(argv):
	(numclusters, filename, sep, timecol, latcol, loncol, catcol) = getArgs()
	splitcsv(filename, numclusters, sep, timecol)
	executor = concurrent.futures.ThreadPoolExecutor(multiprocessing.cpu_count())
	futures = [executor.submit(hostdmp, filename, sep, timecol, latcol, loncol, catcol, i) for i in range(numclusters)]
	concurrent.futures.wait(futures)
	print("The ports used were " + str(ports))


if __name__=='__main__':
	main(sys.argv[1:])