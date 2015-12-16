import json
import requests
import time
"""
def ordered(obj):
    if isinstance(obj, dict):
        return sorted((k, ordered(v)) for k, v in obj.items())
    if isinstance(obj, list):
        return sorted(ordered(x) for x in obj)
    else:
        return obj
"""

def main():
    ports = [29512, 8999]
    count = 0
    t = [0, 0]
    r = [None, None]
    print("Starting benchmark")
    with open('querysample', 'r') as f:
        for query in f:
            count += 1
            query = query.strip()

            for i, port in enumerate(ports):
                # Time the distributed port.
                url = "http://localhost:" + str(port) + query
                start = time.time()
                r[i] = requests.get(url)
                end = time.time()
                t[i] += (end - start)

            # if ordered(r[0].json()) != ordered(r[1].json()):
            #    print(ordered(r[0].json()))
            #    print("\n")
            #    print(ordered(r[1].json()))
            #    print("\n")
            #    print("Query was: " + query)
            #    return
                

    print("Total time for distributed: " + str(t[0]) + "s")
    print("Total time for original: " + str(t[1]) + "s")
    print("Average time for distributed: " + str(float(t[0]/count)))
    print("Average time for original: " + str(float(t[1]/count)))

"""
This is a benchmark tool that we use to test sample requests contained line by
line in a file by running them against our distributed backend (assumed to be
port 29512) and the original nanocube service (port 8999)
"""
if __name__ == "__main__":
    main()
