import sys
import os
import argparse
import concurrent.futures
import multiprocessing
from datetime import datetime

"""
This splitter.py script is a helper script to spawn split a csv file into n
partitions, spawn n servers each serving one partition, and run our master
server on a specified port. Example usage:

    python3 splitter.py -n=5 -f='./nanocube/data/crime50k.csv' \
            -s=',' -t='time' -a='Latitude' -o='Longitude' -c='crime' -p=29512

Will split the csv file into 5 parts and run our master server on port 29512.
"""

ports = []


# parses and returns arguments
def getArgs():
    parser = argparse.ArgumentParser(
        description='Hosts distributed nanocubes by splitting the input .csv file with specified cols and number of clusters')
    parser.add_argument(
        '-n',
        '--numclusters',
        type=int,
        help='Number of clusters',
        required=True)
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
    parser.add_argument(
        '-p',
        '--port',
        type=int,
        help='master server port',
        required=False,
        default=29512)
    # Array for all arguments passed to script
    args = parser.parse_args()
    # Return all arg values
    return args.numclusters, args.filename, args.sep, args.time, args.lat, args.lon, args.cat, args.port

# splitcsv takes a csv filename in the $NANOCUBE_SRC/data and splits it into
# numclusters csv files, each with the same header line
def splitcsv(filename, numclusters, sep, timecol):
    numLines = sum(1 for line in open(filename))
    linesPerFile = numLines / numclusters
    count = 0
    firstNum = 0
    fn = filename.split('.csv')[0] + '_split'
    for line in open(filename, 'r'):
        if (count == 0):
            firstLine = line
        else:
            fileNum = int(min((count - 1) / linesPerFile + 1, numclusters))
            if (fileNum != firstNum):
                f = open(fn + str(fileNum) + '.csv', 'w+')
                f.write(firstLine)
                firstNum = fileNum
            f.write(line)
        count += 1

# returns a random open port
def get_open_port():
    import socket
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.bind(("", 0))
    s.listen(1)
    port = s.getsockname()[1]
    s.close()
    return port

# hostdmp convers the split csv files to dmp files, hosts the nanocube,
# and deletes the original split .csv file
def hostdmp(filename, sep, timecol, latcol, loncol, catcol, i):
    try:
        fileNum = i + 1
        port = get_open_port()
        ports.append(port)
        fn = filename.split('.csv')[0] + '_split'
        dmp_command = "$NANOCUBE_BIN/nanocube-binning-csv --sep='" + sep + "' --timecol='" + timecol + "' --latcol='" + latcol + "' --loncol='" + loncol + "' \
				--catcol='" + catcol + "' " + fn + str(fileNum) + '.csv > '\
            + fn + str(fileNum) + '.dmp'
        os.system(dmp_command)
        host_command = 'cat ' + fn + \
            str(fileNum) + '.dmp | $NANOCUBE_BIN/nanocube-leaf -q ' + str(port) + ' -f 10000 &'
        os.system(host_command)
        rm_command = 'rm ' + fn + str(fileNum) + '.csv'
        os.system(rm_command)
    except:
        print('error with port')


def main(argv):
    if not os.environ.get('NANOCUBE_SRC'):
        print("You must set the NANOCUBE_SRC environment variable before " +
                "running this script.")
        os.exit(1)

    # If the NANOCUBE_BIN environment variable is not set, set it to
    # $NANOCUBE_SRC/bin.
    if not os.environ.get('NANOCUBE_BIN'):
        nanobin = os.environ.get('NANOCUBE_SRC') + "/bin"
        print("Setting NANOCUBE_BIN to " + nanobin)
        os.environ['NANOCUBE_BIN'] = nanobin

    (numclusters, filename, sep, timecol, latcol, loncol, catcol, master_port) = getArgs()
    splitcsv(filename, numclusters, sep, timecol)
    executor = concurrent.futures.ThreadPoolExecutor(
        multiprocessing.cpu_count())
    futures = [
        executor.submit(
            hostdmp,
            filename,
            sep,
            timecol,
            latcol,
            loncol,
            catcol,
            i) for i in range(numclusters)]
    concurrent.futures.wait(futures)
    dist_call = "go run cmd/cli/distnano.go -p " + str(master_port) + " "
    for port in ports:
        dist_call += "-a http://localhost:" + str(port) + " "
    # now run distnano.go using the ports as arguments
    os.system(dist_call)


if __name__ == '__main__':
    main(sys.argv[1:])
