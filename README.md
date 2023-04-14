# An-optimized-2-round-BB

Proof-of-Concept implementation for our improvement on a 2-round Byzantine broadcast protocol. This codebase serves as a tool to test the performance of the optimization presented in the paper (see below). Specifically, it allows for the evaluation of performance under the Byzantine strategy outlined in Section 6.2 on page 11. Additionally, it provides the functionality to commit a value in the optimistic case that the broadcaster is honest and the network is synchronous. 

## Paper 

 <u>Title:</u> An Optimization for a Two-Round Good-case Latency Protocol

<u>Abstract:</u> Byzantine broadcast is a fundamental primitive in distributed computing. A highly efficient Byzantine broadcast protocol, motivated by the real-world performance of practical state machine replication protocols, is increasingly needed. This paper focuses on the state-of-the-art partially synchronous Byzantine broadcast protocol proposed by Abraham et al. (PODC’21). The protocol achieves optimal good-case latency of two rounds and optimal resilience of n ≥ 5f - 1 in this setting. We analyze each step of the protocol, and then improve it by cutting down the number of messages required to be collected and transmitted in the heaviest step of the protocol by about half, without adding any extra cost. This benefits from a new property, named “spread", that we identify and extract from the original protocol. It helps us to eliminate non-essential work in its view-change procedure. We also show that no further reduction is possible without violating the security. We implemented a prototype and evaluated the performance of improved and original protocols in the same environment. The results show that our improvement can achieve about 50% lower communication cost and 40% shorter latency at a scale of 100 replicas. The latency gap becomes wider as the scale further increases.  

## Quick Start

To run the codebase at your machine (with Ubuntu 18.04 LTS) under 4 nodes:

```go
cd TwoRoundBB
go build main.go
```

You can either start each node separately by `./main X`, where X={1,2,3,4}, or start all nodes simultaneously using `bash all.sh`

Next, open a new terminal in the same directory and input `go run start.go` 

## Performance under the Byzantine attack

To simulate the Byzantine strategy outlined in Section 6.2 on page 11 of the paper, do as follows:

1. find the file 'normal_case.go' in the 'consensus' folder;
2. remove line 13, 21, 48, 60, 126, and 132;
3. comment out line 62;
4. rebuild and run the codebase as in Quick Start

## Parameter adjustment

You can test the performance under more nodes. Do as follows:

1. find the file 'nodetable.csv', and append the new node id and IP addresses at the list;
2. go to the file 'genKey.go' for generating signature keys for each node. Assign the total number of nodes to `n` in line 16, and then input `go run genKey.go` at a new opened terminal.
3. go to the file 'parameter.go', and change the assigned value of `N` (in line 5) to be the total number of nodes. Change the value of `F` to any value satisfying N ≥ 5F -1.
4. rebuild and run the codebase as in Quick Start



