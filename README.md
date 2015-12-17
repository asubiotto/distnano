# Distnano
Distributed Nanocubes

12.16.2015
─
Alfonso Subiotto Marqués
Foster Hoff


Introduction
There has been great effort in the last decade to create interactive visualizations of scientific and economic data. Researchers of AT&T laboratories present a novel data structure in their recent publication, Nanocubes for Real-Time Exploration of Spatiotemporal Datasets, which makes use of traditional “cube” and “roll-up” database operations to compress massive spatiotemporal datasets into a data structure which typically fits in a modern laptop’s main memory. 
Motivation
The nanocube data structure allows for interactive visualizations with query response and render times faster than a typical screen refresh rate, however; there are a few major drawbacks of the existing nanocubes backend. Building a nanocube can take hours, defeating the purpose of rapid visualization to gain an initial insight of a dataset. To make things worse, the nanocube can only be appended to, requiring a full rebuild if a chunk of data is deleted. Furthermore, the nanocube does not always fit into a laptop’s main memory for very large datasets. To resolve these major drawbacks,  we set out to distribute the process of building and querying the nanocube data structure so users can leverage the power of multiple machines to speed up the build process. Additionally, if a chunk of data is deleted from the distributed nanocube, only a single node needs to be rebuilt instead of the entire data structure. 
Contributions
1. Implement a distributed system to build and query nanocubes in parallel
2. Benchmark the system to verify correctness of query results and to measure the time taken by the distributed system in comparison to the original backend. 
3. Discuss further steps for improvement of the distributed system and additional feature implementations







Outline


1. Related Work
2. System Architecture + Concepts
3. Implementation
4. Challenges
5. Evaluation
6. Summary
7. Future Work
8. References
9. Appendix















Related Work
Algorithms
In the past two decades, there has been a major effort by the computer science community to develop algorithms and platforms for visualizing multi-dimensional datasets of large scale. One of the simplest of these algorithms is the cube operation, which is simply aggregating data across the time or space dimension in order to summarize that dimension without maintaining all of the data that defines it. Recent work   by Bach et al reviews many recent temporal data visualization techniques based on the space-time cube operation. A useful catalog of their reviews and the corresponding visualizations can be found at http://spacetimecubevis.com/.  A great deal of research has gone into developing other algorithms to explore space and time data while maintaining the dimension of data that the cube operation does not. For example,  M4: A Visualization-Oriented Time Series Data Aggregation, uses a novel min-max approach to aggregate time series for dimensionality reduction without the introduction of any error. Alternatively, Indexing for Interactive Exploration of Big Data Series  presents a way to adaptively index big data series. Instead of indexing the entire data series up front, this technique is used to speed up the process of querying large datasets. Meanwhile, Importance-Driven Time-Varying Data Visualization uses a calculation of entropy to determine the relative information of new data over time. Clustering the importance curves developed by this calculation of entropy has the ability to automatically detect anomalies in scientific data, such as climate data or volcano data. Xi Zhu et al also present a hierarchical clustering technique for “flow data”, which is data with two spatial dimensions, a start and end point. Finally, ScalaR uses a mix of aggregation, filtering, and sampling to dynamically reduce dataset dimensionality for faster query response times. Any traditional DBMS can make use of the techniques implemented by ScalaR. As we can see from this limited selection of algorithm based papers in the last decade, there’s an ample effort being made to effectively visualize and query large datasets to gain meaning from the underlying data.
Systems
There have also been a number of complete systems developed for the ease of querying and visualizing massive datasets. These systems are designed to be used for a wide array of users, from amateur data explorators to more advanced database specialists. For example, PanoramicData: Data Analysis through Pen and Touch by researchers here at Brown University, enables amateur data users to easily visualize data through pie charts, histograms, time-series, and much more. PanoramicData, otherwise known as Vizdom, also enables the quick application of common machine learning algorithms for data exploration. In order to quickly provide results of these algorithms on large datasets, approximate results for a sample of the data are first shown, which are improved over time as the sample grows. Polaris: A System for Query, Analysis and Visualization of Multi-dimensional Relational Databases is an older (2002) system used to visualize data which makes use of the cube operation. ArcGIS is a modern commercial application to build geographic maps of  higher dimensional datasets. As one can see, there are a number of modern platforms designed solely for the visual exploration of large datasets.
Background
The nanocubes data structure is another algorithmic approach to dimensionality reduction for interactive visualization. Nanocubes makes use of the cube operation to aggregate data within a fixed number of time bins to maintain the temporal dimension that other modern cubing algorithms do not. Data is indeed lost by this aggregation, but the size of the time bins can be modified to fit one’s needs. In addition, nanocubes leverages a modern data structure, the quadtree, to enable coarse and fine-grained querying of the spatial dimension, down to the specific geospatial coordinates that defines a data point. This data structure can even be applied to more generic (x,y) data as seen here. Finally, the categorical dimensions are stored as flat-trees so all categories can be maintained as they are summed for data analysis. Nanocubes can support an arbitrary number of categorical dimensions, so, for example, taxi data can be filtered by company, number of passengers, tip amount, or any combination of those variables. The nanocubes project is open-source as it can be found on Github, and it even provides a visualization engine built on top of D3.js and Leaflet.js to conveniently provide visual querying functionality. 

System Architecture + Concepts


General Distributed Architecture
Our system follows the general master/slave communication protocol often found in distributed systems. To help discuss the architecture, let us illustrate an example.


Our user wants to visualize a csv file with 50,000 lines (records) of spatio-temporal data on a cluster of 5 nodes. The user splits the file into 5 csv files each with 10,000 lines and constructs a nanocube on each node with an HTTP interface exposed following the specifications in this file.


On another node (the master), the user starts up our server and passes in the addresses of each of the n slave nodes. At this step, the master performs some initialization steps, and is ready to answer queries following the same API specification by forwarding the requests to each of the n nanocube nodes and merging the results together.


  

An example of how a simple query asking how many records are in the dataset is handled. Note the merging step on the master node.


Queries and Merging
This section will discuss how all queries apart from temporal queries (discussed in the next section due to added complexity) are merged on our master node.


The nanocube system accepts queries through an HTTP interface and answers these with JSON. The simplest type of query is a query with no constraints, as seen in the figure above. The response for a count query (for example) is the simple JSON object:


{ "layers":[  ], "root":{ "val":50000 } }
Where 50,000 is the number of records. The merging process for these kinds of queries is simple as the value contained in root[val] can simply be aggregated across all queries. Other examples of queries following this sort of response style are queries on arbitrary rectangular regions or queries for counts of particular categories.


The next type of queries are queries that require a breakdown of results by quadtree addresses (i.e. diving into the spatial dimension for example) or other attributes. These kinds of queries are answered by JSON objects of the following form:


{ "layers":[ "anchor:location" ], "root":{ "children":[ { "path":[2,1,2,0,0,0,0,1,3,2,3], "val":7606 }, { "path":[2,1,2,0,0,0,0,1,3,1,2], "val":224 }, { "path":[2,1,2,0,0,0,0,1,2,3,3], "val":272 } ]}}
Where rather than root having a direct value, it contains a list of children which is the breakdown of results. In this case “path” refers to a quadtree address and “value” to the result for that particular quadtree address.


Each quadtree address in each of the nanocube nodes refers to the same spatial region since the nanocube specification is that the quadtree represents the spatial dimension. Therefore, the master merges these sorts of queries by aggregating across the same quadtree paths. Another example using different kinds of keys (bidimensional image addresses rather than quadtree locations):


{ "layers":[ "anchor:location" ], "root":{ "children":[ { "x":6, "y":131, "val":36820 }, { "x":7, "y":130, "val":224 }, { "x":6, "y":130, "val":12684 }]}}


These query responses can sometimes also contain another array of “children” for each bidimensional coordinate which is then merged recursively.


Similarly, there are categorical queries whose response follows the structure of this JSON object:


{ "layers":[ "anchor:crime" ], "root":{ "children":[ { "path":[21], "val":2 }, { "path":[20], "val":456 }, { "path":[19], "val":1 }, { "path":[18], "val":1 }, { "path":[17], "val":1 }, { "path":[16], "val":5742 }, { "path":[15], "val":2226 }]}}


These are merged using the same process even though the semantics are different. The query to achieve this result is to ask for a certain breakdown of values by categorical value. Here, each “path” refers to a categorical identifier rather than a quadtree address or a bidimensional image address.


Initialization and Time Bins
This section discusses temporal queries and the added complexity introduced by splitting data with a temporal component into separate nanocubes.


When constructing a nanocube, a user specifies a few attributes that can then be queried by a schema query returning the following:


{ "fields":[ { "name":"location", "type":"nc_dim_quadtree_25", "valnames":{  } }, { "name":"crime", "type":"nc_dim_cat_1", "valnames":{ "CATEGORY0":0, "CATEGORY1":1 } }, { "name":"time", "type":"nc_dim_time_2", "valnames":{  } }, { "name":"count", "type":"nc_var_uint_4", "valnames":{  } } ], "metadata":[ { "key":"location__origin", "value":"degrees_mercator_quadtree25" }, { "key":"tbin", "value":"2013-12-01_00:00:00_3600s" }, { "key":"name", "value":"crime50k.csv" } ] }


This schema response has information regarding the number of possible categories, times, etc… but most important to this section, the schema also refers to a start time and a duration (specified by the tbin key). These are the fundamental notions of time that a nanocube has. Starting from time bin 0 (representing the start time), each time bin represents a certain duration (3600s in this case) and the nanocube is queried according to time bins rather than real time.


Problems then arise when a dataset with a temporal attribute is split and each partition is constructed into a separate nanocube, as each nanocube will then different notions of time. For example, if a dataset includes data ranging from halfway through 2014 to halfway through 2015 and it is split into two, assuming this data is sorted by time, one nanocube’s zeroth time bin will refer to halfway through 2014 while the second nanocube’s zeroth time bin will refer to the start of 2015. The challenge then, is to convert queries over the global notion of time bins into separate queries for each nanocube that takes into account their own time bins.


The initialization step is the way that our master server gets the necessary information from each slave node to then modify time queries as they come in. Information from each node is gathered by a schema query whose time bin component is then interpreted according to a “global” view of time where time bin 0 is the global start time. Each slave node is then assigned a relative bin which corresponds to the time bin where their notion of time starts.


There are two types of temporal queries. The first, or “interval”, query is a simple query over an interval where the start and end time bin are specified:


http://localhost:29512/count.r("time",interval(484,500))
Since these queries are often accompanied by another attribute and they therefore take the form of restricting queries from the previous section to a certain time interval, the merging process shall not be discussed. However, there is a preprocessing step of simply specifying the interval in terms of each slave’s relative start bin. For example, if our slave’s notion of time started at time bin 484, the query issued to our slave would instead be:


http://localhost:29512/count.r("time",interval(0,16))


The second type of temporal query is known as an “mt_interval_sequence” query. This type of query has three parameters: a start bin, how many bins are in a “bucket”, and the number of “buckets” to return. These buckets, or collections of time bins, are simply a resolution desired by the entity issuing the query.


  

Example of an mt_interval_sequence query over “global” time. We want a breakdown of values starting at a and ending at e where bucket size and the number of buckets between the two is specified.


The query is then answered according to buckets:


{ "layers":[ "multi-target:time" ], "root":{ "children":[ { "path":[0], "val":762 }, { "path":[1], "val":724 }, { "path":[2], "val":660 }, { "path":[3], "val":515 }, { "path":[4], "val":410 }]}}


These queries are issued when a graph of a value’s change over time must be constructed, for example.


There is a simple case of this query where our slave node’s notion of time starts at one of the bucket limits and we can simply modify the limits, similar to how the interval query is answered. For example, if a slave node’s relative bin is c from the figure above, we would modify the mt_interval_sequence query to be:


mt_interval_sequence(c,5,2)
This means that we have an additional step when merging. Our slave node in this example will return results in terms of bucket 0 (c-d) and bucket 1 (d-e). However, these are not the correct bucket ids according to our global view of time. A bucket offset is then calculated and applied to bucket ids to convert them back to a global notion of time before a normal merge is executed.


Another case of our mt_interval_sequence query requires us to split a mt_interval_sequence query into two separate queries and merge those together into the result for a slave node before merging that into our global response. This case is when a slave node’s relative time bin is not at one of the bucket limits, but rather, inside a bucket.


The first query sent to our slave node will query the time bins between the relative bin of the slave and the limit of the nearest bucket while the second query will query from the start of the next bucket in relative terms until the end of the global query.


Consider the figure above and let us illustrate the merging of the queries with an example. Let our slave node’s relative time bin be the fourth time bin between b and c. Our first query will query between the node’s relative bin and c and our second query will query from c to e.


When we merge these two queries, the first query can be considered to answer the value in a relative bucket of 0, thus all bucket ids from the result of the second query must be offset by 1. Now that we have this merged query, we must apply the bucket offset from the global query to the bucket ids as in the previous case.

Implementation
Specifications
The nanocube slaves all run the C++ nanocube implementation and our master server is built in golang. Golang was chosen because of its simplicity and ease of use. Additionally, golang has the ability to add “tags” to struct fields to specify JSON decoding/encoding options which made it straightforward to decode and encode JSON into golang structs for manipulation. Here are some type declarations that mirror the JSON query responses shown above:
 Screen Shot 2015-12-16 at 11.05.38 AM.png 

The tags allow you to specify the key whose value should be assigned to the struct field and what key should be given to the value when encoding the struct field into JSON. Tags also allowed us to specify a general structure for the JSON object without having to include all of its fields through the “omitempty” label. For example, sometimes the query responses contain bidimensional image addresses (x and y) and the Path field is omitted.


Apart from the NanocubeResponse shown above, we also have a SchemaResponse struct since each has a different merging strategy.


When our master receives a request, it spawns off a goroutine for each slave node, does any preprocessing necessary, and forwards the request to the slave node. As JSON responses are returned, the master starts building the global response in an O(n) merging process by converting the JSON responses to structs as they come in and merging the first two together, followed with the third and the result of those two etc… The final response is then encoded back into JSON and sent to whoever requested the information.


We also included a python utility file in our codebase named splitter.py. To build a nanocube, one starts with a .csv file, converts that to a nanocube-readable .dmp file with the nanocube-binning-csv utility provided by the C++ source code and finally executes the nanocube-leaf program to build the nanocube from the .dmp file and start serving queries on a specified port.


Our splitter.py utility takes care of all these steps for the user but is only for running slave nodes on one machine as it splits the .csv file into a specified number of partitions, converts those partitions into .dmp files, and builds a nanocube over every partition serving queries on open ports on localhost followed by spawning the master server. This utility was useful to us in testing and automating the splitting and building process.
Example Run
We will do an example run of our system with a Chicago crime dataset with 50,000 records and 4 nanocube slaves.


The first step is to export our environment variable to the C++ nanocube source code so that splitter can find the utilities necessary to build nanocubes:
 Screen Shot 2015-12-16 at 11.28.06 AM.png 

We then run splitter.py on the .csv file to initialize our system:
 Screen Shot 2015-12-16 at 11.25.01 AM.png 

In this command we specify -n=4, the number of slave nodes, the path to the .csv file and some .csv specific information. For example -s is the separator while the other arguments specify the name of the temporal, spatial, and categorical columns. Finally, -p is the port on which to spawn the master server.


Our system is now up and running with the master server running on port 29512. The next step is to start the visualizer to view our data as specified in the C++ nanocube implementation README by first building a configuration file for the data:
 Screen Shot 2015-12-16 at 11.31.17 AM.png 

We can then run the visualizer on a certain port by changing directory to “$NANOCUBE_SRC/extra/nc_web_viewer” and running:
 Screen Shot 2015-12-16 at 11.32.52 AM.png 

Where 8000 is the port on localhost where we will access the visualizer. To view our data we point our browser to “localhost:8000/#config_crime” where config_crime is the name of the configuration file generated in the ncwebviewer-config step and we can view our data with our distributed system as the backend.


 Screen Shot 2015-12-16 at 11.35.04 AM.png 




Challenges
Datasets
We faced a number of issues in tailoring the existing nanocubes distribution to our needs. To test the correctness of our system, we had to first find a variety of different dataset that meet the specifications of the open source project. Each dataset must have latitude and longitude, date-time, and a categorical variable for testing. While having these three factors isn’t that uncommon, we needed this dataset to be large and publically available, greatly limiting our options. To make matters worse, a categorical variable can’t be a quantitative variable like temperature, disabling the use of many scientific datasets, such as those published by NASA. As an aside, it is possible to turn a quantitative variable into a categorical variable by binning it, but this was beyond the limitations of our project. Furthermore, the existing code base only supports building nanocubes from .csv file types. We tried to find the datasets published in the benchmarks of the research paper, but only three of the datasets were publically available, and they required a significant amount of cleaning to meet the nanocubes blueprint. It then seemed that there was more to meet the eye than we had first expected from the elegant publication.
Date-Time
After we finally found a few datasets that met the specifications of the nanocubes distribution, we ran into some unforeseen errors. To build a nanocube from the data, the script nanocube-binning-csv must be run to convert a .csv file into a .dmp binary file. It was assumed that this script would work out of the box, but there was an error in converting certain date-time formats that we had to fix.  We’ve requested to merge our changes with the open source code and hope to see this fix implemented for the public. This wasn’t the only issue we faced with the time data, however. When we tried to build our first original nanocube from the San Francisco crime dataset, everything seemed to be going as planned, until there was a major spike in the time spent on the build process. It was taking about three seconds per ten thousand entries to append to the nanocube, then suddenly at about one million entries, it started taking forty seconds to add each set of ten thousand entries. We were baffled at first. Fortunately, Lauro Lins, one of the primary authors of the nanocubes research project, was quick to respond by email. He first noted that it seemed like the time data was not sorted properly. He said, “If the timestamp of the current record is smaller than a previously stored record in that same multidimensional bin, then we need to open a slot in the middle of the vector, push elements to the right and recompute cumulative values stored in that vector. This adds an extra linear cost in the length of the time series instead of constant time.” This additional linear cost made sense based on the way time is defined to be stored by the research paper, “as a dense sorted array of cumulative counts tagged by timestamps,” but we discovered that the conversion from .csv to .dmp sorts the data by time, so this could not be the issue. After more communication, Lauro made the astute observation that there are only 216  time bins by default. Since we were building a nanocube of San Francisco crime data since 2003 with one hour time bins, we would have required 105,120 time bins which is far greater than 216  . The more recent times then proceeded to wrap around to the beginning of the time bins, resulting in the linear insertion time cost that he had previously mentioned. Increasing the time resolution to four hours resolved this issue, but it could have also been resolved by increasing the number of time bins. 
Memory limitations
One final issue we faced is that we were limited to machines with only 8 gigabytes of total RAM. When building our largest dataset, the New York taxi dataset, our system began pushing the nanocube to disk at 4555 megabytes, which occurred at 35,380,000 entries. The build process then began to take longer and became less predictable, requiring we terminate our benchmarks at this point when the system flushes the nanocube from memory to disk. This is a major drawback because the original nanocubes research paper shows that key saturation begins to occur at around this point of 50,000,000 entries. Key saturation means that fewer unique keys are found over time, so exponentially less memory is required to build the nanocube over time. We began to see less memory being required as our nanocube approached this point of 35,380,000 entries, but not nearly to the exponential degree described in the research paper for the tests with one billion entries. The implication of this limitation is that we could not extensively compare the memory required by our distributed system for a gargantuan dataset with the memory required to host the nanocube on a single node. We would expect to see far less key saturation for each distributed node, as different nodes may share keys. To account for this, we propose an intelligent data partitioning scheme based on the structure of the quadtree in the future work section of the report, which would cause fewer keys to be shared between different nodes. 

Evaluation
Benchmarks: Nanocubes build process
To evaluate the performance of our distributed building process in comparison to the original build process for a single nanocube, we had to first find some larger datasets to work with. Finding large publically available datasets that met the exact specifications of nanocubes was a challenge, but we found three sets to work with. We ended up selecting the full Chicago crime dataset, San Francisco crime dataset, and New York taxi dataset. The reason why it’s necessary to benchmark the build process on a large dataset is because nanocubes are inherently designed to be more efficient with large datasets, as the key space begins to visibly saturate around 50-100 million entries. The New York taxi dataset consisted of over 170,000,000 records, but unfortunately we could not test this many due to lack of main memory. As we can see from the plot below, of memory required per number of entries for the taxi data, the logarithmic curve of memory growth begins to take shape even after only 35,380,000 records. The time required to append to the nanocube clearly continues to grow linearly as the number of records grows intensely.


 Taxi.png 
    
Nanocube memory usage (left) and build time (right) per number of records for a single node using the existing backend on 35,380,000 records of the taxi dataset. The memory usage begins to take a logarithmic shape as the number of records grows whereas the insertion time remains linear.


We expect the saturation of the key space to be a major downfall for the distributed version of the nanocubes, for each partition may share a large portion of the key space. One way to avoid this issue is to partition the data in a manner such that the nanocubes share as little of the key space as possible, which we discuss in the future work section of this report. 
For the smaller datasets, Chicago crime (5,899,717 records) and San Francisco crime (1,848,216 records), the logarithmic scaling of memory due to key space saturation is not  obvious. We tested each of these datasets, as well as the taxi data, for 1, 2, 3, and 4 nodes to see how much time and memory each fraction of the data would take to build. To be clear, we only tested the build time of a fraction of the dataset using the sequential partition scheme, since each node should take approximately equal time and memory with equal sized partitions. We found that the build time linearly decreases as the number of nodes increases. The memory required also linearly decreases as the number of nodes increases, even for the taxi data, since it isn’t large enough to clearly see the key space saturate.


 ntdmu.jpg 
  

New York taxi dataset memory usage (left) and build time (right) per number of nodes (where the amount of data per node is the number of records divided by the number of nodes). 
 sfcdmu (2).jpg 
 sfcdbt (2).jpg 

San Francisco crime dataset memory usage (left) and build time (right) per number of nodes (where the amount of data per node is the number of records divided by the number of nodes). 


 ccdmu.jpg 
 ccdbt.jpg 

Chicago crime dataset memory usage (left) and build time (right) per number of nodes (where the amount of data is the number of records divided by the number of nodes). 
Benchmarks: Nanocubes query process
We used the Chicago crime dataset with 50,000 records and a size of 13MB to benchmark the average query time against the original nanocube implementation. The important information to be gained was how the average query time for an average workload with our distributed system as its backend differs from the average query time for the same workload with the original nanocube backend.


To this end, we ran around 600 queries sampled from a normal workload (i.e. queries sent from the visualizer to the nanocube) against both our distributed backend and the original nanocube backend and got an average response time as n (the number of nodes in our distributed backend) is increased:


 distquery.jpg 

The change in average query time for our distributed system as n is increased.
It is important to note that the x axis is in terms of log2(n). We can see that, starting at 8 nodes, the average query response time for our system starts to grow linearly as the number of nodes is increased because the merging process starts to become a bottleneck.


Even with this linear increase in average query response time, the original nanocube paper prides itself in returning queries faster than the average screen refresh rate (60fps = 16.7ms per frame) and we can note that our system starts to deliver queries a bit slower than the screen refresh rate at 32 nodes for this set of queries.


The results shown above show that our nanocube backend can still achieve real-time visualization of large spatio-temporal data for a certain number of nodes. However, the query latency could be tolerated by a user since it means that as n increases,  the total time to build the dataset decreases. The optimal balance is finding a cluster size that both massively reduces the total build time yet still achieves response time faster than the average screen refresh rate.


The average query response time can further be reduced by modifying the merge process. The merge process is currently a naive O(n) implementation which can easily be extended to an O(logn) implementation. This approach is discussed more in the future work section.





Summary


When we started this project, we wanted to solve the problem of lengthy nanocube build times as well as too much memory usage for very large spatio-temporal datasets while keeping the visualizations of this data as close to real-time as possible.


We went about solving this problem by seeing how feasible it was to split a dataset into smaller partitions controlled by a master that forwards queries to the slaves and then merges these.


The process of splitting the data into several partitions and intelligently aggregating queries over the partitions turned out to produce the same results as querying one large nanocube over the whole dataset because the nature of the nanocubes data structure is to produce answers for queries on value aggregates. Thus splitting values into subsets and aggregating the subset aggregates produces the same results as aggregating the original values.


With this in mind, all the queries were straightforward to aggregate over with some additional work needed for temporal notions. Our system resulted in achieving the same latency as a single node system up to a certain number of nodes, where the merging process started to become a bottleneck. However, our system still achieves near real-time query results and is a feasible alternative to a single-node nanocube system when a nanocube over a dataset is simply too large to fit into memory or takes too long to build.

Future Work
O(n) to O(logn) merging process
As the evaluation of the query process mentions, the average query time for our distributed system can be improved from following an O(n) trend to following a best-case O(logn) trend.


The merging process is naive: as query responses are received, the master server keeps around a global response to be returned to the issuer of the query and merges each returned response into the global response sequentially.


A way to improve this merging process would be to follow a tree-based approach to merging where a tree has as many leaf nodes as half the number of slave nodes. This means that as responses are received, the responses are assigned to a leaf node to be merged. Once the number of responses in a leaf node reaches 2, the responses are merged and propagated to the parent node and the process is repeated. Responses are not pre-assigned but rather dynamically assigned to leaf nodes that have not been used and that preferably have one occupant.


The worst-case runtime of this strategy is the case where the responses are received sequentially and no merges can take place simultaneously so O(n) but the best case is when all the merges on a level of a tree take place simultaneously leading to an O(logn) runtime.
Fault tolerance
We have not implemented fault-tolerance for our current version of the system but doing so should be pretty straightforward.


The current implementation of nanocubes does not allow for the removal of records from the nanocube but does allow for the addition of records in real-time.


The idea for a fault-tolerant system would be to replicate every slave node 3 or 5 times and have each slave node cluster follow the Raft consensus algorithm so that additions can be agreed upon. A consistent nanocube can then be maintained even with the failure of a few slave nodes in a cluster.
Quadtree partitioning for benefits with memory usage, insertion speed, query time 
Below is an image detailing the nanocube build process for an example with five objects (tweets) with associated spatial, temporal, and categorical dimensions:


 Screen Shot 2015-12-17 at 10.11.13 AM.png 



We can see from the graphical representation of the nanocube that the quadtree spatial dimension is the lowest level of the tree. Quadtrees are often used to store more general spatial data, so research has been done to figure out a way to intelligently partition spatial data for distributed quadtrees. This research shows that the partitioning scheme can greatly speed up the process of building and merging quadtrees. The goal of the scheme is to ensure that each tree stores a mutually exclusive portion of the tree so there is no overlap of the query and build process.


As mentioned before, due to the naive partitioning of the csv file into chunks, the slave nodes will have quadtrees that share a large portion of the key space. By manipulating the quadtree partitioning scheme so none of the spatial data overlaps, we believe we can ensure that keys are mostly unique to each node’s nanocube. This also implies that given a spatial query, only a fraction of the nodes will need to be queried depending on whether they contain those spatial coordinates or not. However, using this partitioning scheme will require a preprocessing step which may require sorting the data by coordinates, or a linear scan at the very least, to determine the fraction of data in each quadrant of the quadtree. A full data scan can be costly for a massive set of data, but the tradeoff is for a lower memory requirement by the overall reduction of the key space. We believe that this quadtree partitioning scheme will allow for key space saturation of the distributed nanocubes, just as originally described in the nanocubes research paper. 



References


GitHub for quadtree partition scheme
Direct communication with Lauro Lins by email
Nanocubes GitHub repository
Nanocubes for Real-Time Exploration of Spatiotemporal Datasets
A Review of Temporal Data Visualizations Based on Space-Time Cube Operations 
http://spacetimecubevis.com/.  
 M4: A Visualization-Oriented Time Series Data Aggregation
Indexing for Interactive Exploration of Big Data Series 
Importance-Driven Time-Varying Data Visualization
Mapping Large Spatial Flow Data with Hierarchical Clustering
 Dynamic Reduction of Query Result Sets for Interactive Visualization
Chicago crime dataset
San Francisco crime dataset
New York taxi dataset

Appendix
Github
https://github.com/asubiotto/distnano
Division of labor
Alfonso handled building the distributed backend in Golang, using his experience from Distributed Systems. He also handled verifying the correctness and benchmarking the distributed query process.
Foster handled dataset acquisition and cleaning, using his experience from Data Science. He also built the Python partitioner which splits a .csv file into multiple files with the same header, converts them into .dmp files using nanocube-binning-csv, and spawns each nanocube from the .dmp files before communicating the location of these nanocubes to the GoLang backend.
