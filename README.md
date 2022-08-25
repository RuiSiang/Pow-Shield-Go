# WIP: PoW Shield Go

Implementation of PoW Shield in Go for stress-testing purposes

## Usage

### Install the package

```bash
go mod tidy
```

### Run

```bash
go run main.go
```

## Stress Test

Note: This only works on non-containerized version of PoW Shield, and that your system might experience unstability when running the test.

```bash
# Start the stress test
npm run stress

# If you changed the PORT variable in .env, you should also change the target variable in the stress test script
nano scripts/stress.sh
```

_The following tests are are conducted on i7-12700H CPU with a sum of 1 100% utilized core and a 60 second period for each concurrent parameter._

### Mass GET

| Concurrent Connections | Avg Latency | Error Rate | Requests/Second |
| ---------------------: | ----------: | ---------: | --------------: |
|                     64 |          ms |          0 |                 |
|                    128 |          ms |          0 |                 |
|                    256 |          ms |          0 |                 |
|                    512 |          ms |          0 |                 |
|                   1024 |          ms |          0 |                 |
|                   2048 |          ms |          0 |                 |
|                   4096 |          ms |          0 |                 |

### Nonce Flood

| Concurrent Connections | Avg Latency | Error Rate | Requests/Second |
| ---------------------: | ----------: | ---------: | --------------: |
|                     64 |          ms |        N/A |                 |
|                    128 |          ms |        N/A |                 |
|                    256 |          ms |        N/A |                 |
|                    512 |          ms |        N/A |                 |
|                   1024 |          ms |        N/A |                 |
|                   2048 |          ms |        N/A |                 |
|                   4096 |          ms |        N/A |                 |
