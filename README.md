# sysinfo

sysinfo is a standalone, read-only Linux and macOS CLI for collecting the system facts needed
to build, distribute and troubleshoot native applications. It gathers OS, kernel,
CPU, memory, libc and process-limit information into a single JSON report, so users
can compare environments, choose suitable build targets and investigate startup failures.

Enhanced from [zcalusic/sysinfo](https://github.com/zcalusic/sysinfo), it adds runtime
compatibility information while keeping the collector independent of installed
language runtimes and diagnostic utilities. For example, when Doctor cannot start,
sysinfo can still collect the host facts needed to investigate its build requirements.

Run it without arguments to print the report; redirect stdout to save it.

```sh
./sysinfo-1.1.4-linux-arm64 > sysinfo.json
# On Linux x86-64:
./sysinfo-1.1.4-linux-amd64 > sysinfo.json
# On Apple Silicon macOS:
./sysinfo-1.1.4-darwin-arm64 > sysinfo.json
```

The binary needs no Bun, Node, Python or shell utilities. Linux builds are static;
macOS builds use only system libraries shipped with macOS.
It does not execute subprocesses or modify system files. Root is optional;
firmware/serial information may be unavailable without permission.

## Information collected

On Linux:

- OS distribution, kernel and userland architecture, CPU and common feature flags.
- Hardware inventory inherited from upstream: firmware, memory, storage and network.
- Actual base page size in bytes (including 64 KiB ARM64 systems), independently of huge pages.
- Installed libc candidates: resolved path, ELF machine, release banner and exported ABI versions.
- This process's soft/hard resource limits, including virtual address space and core dumps.
- `/proc/meminfo`, overcommit policy/ratio/size, mapping limits and cgroup membership.

On macOS, native system APIs collect the OS version/build, Darwin kernel, host and
process architectures, CPU model/core counts, physical memory, base page size,
resource limits and network interfaces. Linux-specific libc, overcommit and cgroup
fields remain empty, with platform notes in the report. Firmware/storage inventory
and ARM CPU feature flags are not collected on macOS.
Under Rosetta, the report identifies translation and separates the ARM64 host
from the AMD64 process. Page size and resource limits describe the running process.

Reports describe the environment visible to the collector. Run it from the same
shell/container/service context as the failing program when investigating limits.
Memory data comes from procfs; cgroup membership is included as context, but cgroup
memory ceilings are not calculated. Host memory is not a container's effective allowance.

`runtime.errors` records unavailable runtime facts; per-library errors are recorded
in `runtime.libc[].error`. Older inventory fields follow upstream's best-effort
behavior and omit unavailable values. A report with gaps is still printed successfully.
Installed libc is searched in standard `/lib`, `/lib64`, `/usr/lib`, `/usr/lib64`
and multiarch subdirectories. Custom prefixes may not be found. `GLIBC_*` ABI tags
are separate from release versions; a missing release banner stays unknown.

This is evidence collection, not an automatic compatibility verdict or crash debugger.
Reports include hostnames, machine IDs, serial numbers and network identifiers;
review them before sharing outside your organization.

## Build and verify

Requires Go 1.24 or later. No third-party Go modules are used.

```sh
make fix
make lint
make test
make build
# dist/sysinfo-1.1.4-linux-amd64
# dist/sysinfo-1.1.4-linux-arm64
# dist/sysinfo-1.1.4-darwin-amd64
# dist/sysinfo-1.1.4-darwin-arm64

# Optional: build a single architecture, or choose an output directory.
make build-linux-arm64
make build DIST_DIR=/tmp/sysinfo-dist
```

`make` and `make build-all` also build all four platform binaries. Artifact versions come
from `version.go`, matching the version in the JSON report. AMD64 uses the v1 CPU
baseline; ARM64 uses ARMv8.0. Both Linux builds use `CGO_ENABLED=0`, so there are no
separate Debian/Kylin, glibc or kernel-labelled binaries. Go's documented Linux
kernel requirements are listed in [Minimum Requirements](https://go.dev/wiki/MinimumRequirements);
the collector also needs access to procfs/sysfs for the corresponding fields.

All four builds use `CGO_ENABLED=0`. macOS binaries use native sysctl and resource-limit
APIs without invoking external commands. Kernel/distribution and page-size combinations
still require runtime validation on those systems.

## Origin and license

Based on [zcalusic/sysinfo](https://github.com/zcalusic/sysinfo), copied from commit
`64129099fd651e9158c261fbf5afc2dd92961a07` (upstream 1.1.3). This CLI starts at 1.1.4.
Upstream's MIT license and copyright notices are retained; `cpuid` retains its BSD license.
