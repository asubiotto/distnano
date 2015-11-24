# Distnano
The aim of this project is to provide a golang wrapper for [nanocube](https://github.com/laurolins/nanocube) so that instead of building one huge nanocube for a dataset, you can spawn n nodes that each build 1/nth of the nanocube and respond to queries over that 1/nth.
The advantages of this project are:
* The process of building the nanocube can be parallelized over n nodes, massively reducing the nanocube build time.
* The API stays the same for client applications since our golang server takes care of querying and merging the results from n nodes.
