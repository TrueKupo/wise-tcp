## Overview: Wise TCP Server

This project implements a **TCP-based Proof-of-Work (PoW) system** with a **client-server model**. The server verifies
that a client solves a computational PoW challenge before granting access to random quotes. The system is **modular**,
lightweight, and Dockerized, designed to demonstrate secure TCP communication with PoW protection against abuse.


## Running / Testing

### Using Docker:

1. **Run the server**:

```
   docker build -f docker/server.Dockerfile -t wise-tcp-server .
   docker run -p 9001:9001 -e PORT=9001 -e MAX_CONN=2 -e POW_DIFFICULTY=20 wise-tcp-server
```

2. **Run the client**:

``` 
   docker build -f docker/client.Dockerfile -t wise-tcp-client .
   docker run -e SERVER_ADDR=host.docker.internal:9001 -e TRY_REPLAY=false wise-tcp-client
```

### Using scripts:

1. **Run the server**:

```
   sh build-server.sh
   sh run-server.sh
```

2. **Run the client**:

``` 
   sh build-client.sh
   sh run-client.sh
```

### Manually (without Docker):

1. **Server**:
    - Build: `go build -o server ./cmd/server/server.go`
    - Run: `./server`

2. **Client**:
    - Build: `go build -o client ./cmd/client/client.go`
    - Run: `./client <server_address>`

Configuration files are located in the `cfg/` directory.


## Design and Functionality

### 1. **Core Concept: Proof-of-Work (PoW)**

- Uses a **Hashcash-like PoW mechanism** to prevent abuse.
- Clients must solve a computationally challenging problem issued by the server before receiving a random quote.

### 2. **Components**

1. **Client**:
    - Requests random quotes from the server over a TCP connection.
    - Solves PoW challenges provided by the server.
    - Supports replay attack testing via configurable settings (`TRY_REPLAY`).

2. **Server**:
    - Accepts multiple client connections over TCP.
    - Issues PoW challenges for client verification.
    - Provides random quote after successful PoW validation.
    - Limits the number of active connections and supports graceful shutdown.

3. **Supporting Modules**:
    - **PoW Library** (`internal/pow/hashcash`): Handles challenge generation, verification, solving, and replay
      protection.
    - **Quote Handler** (`internal/handler`): Retrieves random quotes from the ZenQuotes API or uses local fallback
      quotes.
    - **Graceful Shutdown** (`internal/graceful`): Ensures smooth resource cleanup during shutdown.
    - **Configuration and Logging** (`pkg/config`, `pkg/log`): Manages YAML configuration and structured logging.


## Implementation Details

### Chosen Algorithm: **Interactive Hashcash**

Employs a **synchronous challenge-response** approach where the connection remains open while the client computes a
  valid token.

### Implementation Breakdown

- **Configurable Work Factor:** The computational difficulty is pre-configured and can be adjusted as needed.
- **Pre-Computation Resistance:** The challenge includes the client’s **IP address, port, and a random nonce**,
  preventing attackers from solving challenges in advance.
- **Time-Constrained Challenges:** The system enforces expiration times to prevent delayed replay attacks.
- **Replay Attack Protection:** Spent challenges are tracked in-memory to prevent reuse.

### Limitations & Current Challenges

- Requires an **open connection** while waiting for the PoW solution (partially mitigated by **connection timeouts**).
- **Static difficulty level:** Does not yet support dynamic adjustment based on real-time server load.

### Areas for Improvement

- **Integration of metrics collection** for **profiling and monitoring**.
- **Enhanced tracing and logging** for issue investigation.
- **Expanded test coverage** to improve robustness and reliability.
- **Refactoring key design components** for improved flexibility and modularity.


## Protocol and Algorithm Selection Rationale

### Protocol Types

#### 1. Challenge-Response (Interactive Proof-of-Work)

**Description:** The server issues a challenge requiring computational work before granting access.

**Advantages:**

- Ensures fair allocation of computational effort.
- Dynamically adjustable workload based on server load.
- Enables graceful degradation under high loads.
- Resistant to pre-computation attacks due to client-specific challenges.

**Disadvantages:**

- Susceptible to connection-slot exhaustion attacks.
- Potential vulnerability to backlog saturation.

#### 2. Precomputed Token Verification (Non-Interactive Proof-of-Work)

**Description:** Clients pre-compute a token that the server verifies.

**Advantages:**

- Stateless server implementation, reducing memory footprint.
- No persistent connection required.

**Disadvantages:**

- Vulnerable to pre-computation attacks.
- Fixed difficulty, cannot adapt to server load.
- Requires replay attack prevention mechanisms.

### Cost Function Types

#### 1. CPU-Bound

**Description:** Computationally intensive, relying on CPU speed.

**Advantages:**

- Simple to implement.
- Relatively predictable solving times.

**Disadvantages:**

- Low ASIC resistance.
- Susceptible to hardware acceleration.

#### 2. Memory-Bound

**Description:** Memory access intensive, relying on RAM speed.

**Advantages:**

- Higher ASIC resistance than CPU-bound.

**Disadvantages:**

- Lower ASIC resistance than Probabilistic Unbounded.
- Less predictable solving times.
- Still susceptible to botnets and distributed attacks.

#### 3. Probabilistic Unbounded

**Description:** Computationally intensive with added randomness.

**Advantages:**

- High ASIC resistance.
- Can be used for adaptive difficulty scaling.

**Disadvantages:**

- Highly variable and unpredictable solving times.
- Susceptible to hardware acceleration.

### Interactive vs. Non-Interactive PoW

#### Interactive PoW

- **Advantages:** Dynamically adjustable workload, resistant to pre-computation attacks.
- **Disadvantages:** Requires an open communication channel, susceptible to connection-slot depletion.

#### Non-Interactive PoW

- **Advantages:** No persistent open connection required.
- **Disadvantages:** Vulnerable to pre-computation attacks, fixed difficulty.

### Conclusion & Selected Approach

For DDoS protection, **Interactive PoW with a carefully chosen cost function** is the optimal approach. Specifically, a
**Hashcash-based challenge-response mechanism** offers the best balance of security and practicality.

**Rationale:**

- **Interactive nature:** Allows for dynamic difficulty adjustments based on server load, mitigating attack intensity.
- **Challenge-response:** Prevents pre-computation attacks by tying challenges to client-specific parameters.
- **Hashcash:** Provides a relatively predictable cost function (low variance) ensuring fair work distribution and
  mitigating the risks associated with probabilistic unbounded functions. It is also resistant to parallelization,
  limiting the effectiveness of distributed attacks.

**Key Considerations:**

- **Connection-slot exhaustion:** Must be addressed through techniques like horizontal scaling and gateway
  optimizations.
- **Cost function selection:** Fixed-cost or low-variance functions are preferred for fair work distribution. Avoid
  parallelizable functions to limit distributed attack effectiveness.


## Future Roadmap

- Implement **adaptive PoW difficulty scaling** based on live server load and attack patterns.
- Integrate **monitoring tools**, including **real-time metrics and tracing**.
- Expand test coverage with **unit and stress tests**.
- Improve connection management strategies to **mitigate connection-slot depletion risks**.
