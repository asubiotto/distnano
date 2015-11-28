import sys
import os


#splitcsv takes a csv filename in the $NANOCUBE_SRC/data and splits it into 
#numClusters csv files, each with the same header line
def splitcsv(filename, numClusters):
	numLines = sum(1 for line in open(filename))
	linesPerFile = numLines / numClusters
	count = 0
	firstNum = 0
	for line in open(filename, 'r'):
		if (count == 0):
			firstLine = line
		else:
			fileNum = min((count-1)/linesPerFile + 1, numClusters)
			if (fileNum != firstNum):
				f = open(filename.split('.csv')[0]+'_split' + str(fileNum) + '.csv', 'w')
				f.write(firstLine)
				firstNum = fileNum
			f = open(filename.split('.csv')[0]+'_split' + str(fileNum) + '.csv', 'a')
			f.write(line)
		count = count + 1
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
def hostdmp(filename, numClusters):
	ports = []
	for i in range(numClusters):
		fileNum = i + 1
		port = get_open_port()
		ports.append(port)
		dmp_command = "nanocube-binning-csv --sep=',' --timecol='time' --latcol='Latitude' --loncol='Longitude' \
				--catcol='crime' " + filename.split('.csv')[0]+'_split' + str(fileNum) + '.csv > '\
				 + filename.split('.csv')[0]+'_split' + str(fileNum) + '.dmp'
		os.system(dmp_command)
		host_command = 'cat ' + filename.split('.csv')[0]+ '_split' + str(fileNum) + '.dmp | nanocube-leaf -q ' + str(port) + ' -f 10000 &'
		os.system(host_command)
		rm_command = 'rm ' + filename.split('.csv')[0]+'_split' + str(fileNum) + '.csv'
		os.system(rm_command)


def main():
	print('Please specify the csv filename') #not the full path, must be in the same directory
	filename = sys.stdin.readline().split('\n')[0]
	print('Thanks. Now how many clusters are you using?')
	numClusters = int(sys.stdin.readline().split('\n')[0])
	splitcsv(filename, numClusters)
	hostdmp(filename, numClusters)


if __name__=='__main__':
	main()