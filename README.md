# MapReduce Implementation

This repository contains a simple MapReduce framework along with example coordinator and worker applications to demonstrate its core functionality.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Makefile Targets](#makefile-targets)
  - [Build](#build)
  - [Run](#run)
  - [Tests](#tests)
  - [Clean](#clean)
- [Example Usage](#example-usage)

---

## Prerequisites

- Go 1.20+ with plugin support
- GNU Make

## Makefile Targets

The `Makefile` provides convenient, reusable commands to build, clean, and run the MapReduce framework.

### Build

- **build/plugins**:
  Compile all MapReduce application plugins (`.so` files) in `mrapps/`. Automatically cleans existing plugins first.

- **build/seq**:
  Build the sequential driver (`mrsequential`) in `main/`.

- **build/mrcoordinator**:
  Build the distributed coordinator executable.

- **build/mrworker**:
  Build the worker executable.

- **build/mr**:
  A convenience target that runs `create/mr-tmp`, `build/plugins`, `build/mrcoordinator`, and `build/mrworker` in sequence.

```sh
make build/mr
```

### Run

- **run/seq**:
  Execute the sequential MapReduce driver on the Word Count plugin and sample inputs.

- **run/mrcoordinator**:
  Start the coordinator with sample inputs.

- **run/mrworker**:
  Launch a worker process against the Word Count plugin.

```sh
make run/seq
make run/mrcoordinator
make run/mrworker
```

### Tests

- **run/tests**:
  Run the full integration test suite (`test-mr.sh`) in `main/`.

```sh
make run/tests
```

### Clean

- **clean/out**:
  Remove all files in `main/mr-tmp/`.

- **clean/plugins**:
  Remove all compiled plugin (`.so`) files from `mrapps/`.

```sh
make clean/out
make clean/plugins
```

## Example Usage

Build everything and run the sequential example:

```sh
make build/mr
make run/seq
```

The output will appear in `main/mr-tmp/`, showing intermediate and final files for each MapReduce phase.

To try a distributed run in separate terminals:

```sh
# In terminal 1: start coordinator
make build/mr
make run/mrcoordinator

# In terminal 2: start worker
make run/mrworker

# (Optional) start multiple workers
make run/mrworker & make run/mrworker &
```

> This project is based on [MIT 6.5840: Distributed Systems](https://pdos.csail.mit.edu/6.5840/). Refer to the course site for full lab details and documentation.
