# My Own Cluster

This is an experiment to provide a full blown yet scalable, durable, simple and flexible computing platform based on new standards and able to run and integrate legacy POSIX applications.

It is designed from the beginning to be simple to operate, especially disaster recovery is easy and always possible. Backups is managed by the software itself. Giving an instance the backuped storages puts the environment in production again.

Elements :

- WASM
- WASI
- KVM
- WebGPU
- ...

For a computer to be useful, it needs a model for :

- Computing,
- Storage,
- Network and IO.

## Building and running

Clone this repository :

```bash
git clone git@github.com:ltearno/my-own-cluster.git

cd my-own-cluster
```

The first time, download the dependencies :

```bash
make build-prepare
```

Then to run it, call this :

```bash
# Some values will be asked for https autosigned certificates generation,
# you can type [ENTER] until the end to use default values
make build run-serve
```

The program should build and start. It will listen on port 8443 with https protocol.

### Testing

In another terminal, go to the repository directory and type :

```bash
# to build the executable binary (the previous step used "go run")
make build

# the samples directory
cd _samples

# this will upload some webassembly compiled sources to the service instance
make register

# this will call the uploaded code and show some results
make call
```

## Executable formats

My-own-cluster executes :

- `web assembly` byte code thanks to the [wasm3](https://github.com/wasm3/wasm3) interpreter,
- `javascript` code thanks to the [duktape](https://github.com/svaarala/duktape) interpreter (there is no nodejs runtime).
- `glsl` OpenGL shader language, by using OpenGL NVidia/AMD drivers loaded by EGL.

Each function will be executed in a jailed and sandboxed very light VM instance (requires KVM, not yet fully implemented, code in [experiments/kvm](experiments/kvm/)).

The functions can interact with the platform through the Guest API which is described just below.

## Exchange Buffers

Exchange buffers is a concept merging files, http requests and responses, web sockets and so on...

In memory it looks like this (data structure)

[ EXCHANGE BUFFER ]
Lock :          mutex // a lock to access the buffer
Headers :       map[string]string
StatusCode :    int
WriteFinished : boolean     // if closed, an exchangebuffer will not accept new parts. Once closed, it cannot be reopened
Parts :         [][]byte    // a queue of not consumed parts
One only reader at a time
One only writer at a time

Usually an exchange buffer is manipulated through two interfaces serving two different purposes : a reader and a writer.

TODO : tranform ExhcnageBuffer to be a memory data structure
TODO : a http request creates a input buffer (closed if http, still opened if websocket)
TODO : request processing finishes when the output buffer is closed

## Guest API

From the _guest_ to the _my-own-cluster_ host.

To support legacy POSIX applications, the host implements most of its API through WASI. There is also a direct binding for applications not using `libc`. _partially implemented_.

File system :

- `api://input` : application input payload
- `api://output` : application output payload
- `http://` and `https://` : used by the guest application to issue a request to some JSON REST Service services.

There is also an API that my-own-cluster provides for guests. It does not require POSIX at all.

The bindings are available in :

- `rust` : in this file [assets/my-own-cluster-guest-api.rs](assets/my-own-cluster-guest-api.rs)
- `C/C++` : in this file [assets/my-own-cluster-guest-api.rs](assets/my-own-cluster-guest-api.rs)

Those functions can be called by whatever language which is targetting WASM and has an FFI (all serious ones).

## Core REST API

A basic set of REST api is provided through the _core_ module. 
This module is loaded at startup like other functions and use the guest api to provide the REST API that the _cli_ program uses.

The core api module implementation can be found in the [assets/rest-default-api.js](assets/rest-default-api.js) file.

## Automatic module binding

You can import a wasm module and my-own-cluster will bind a stub to module registered with same name if it exists. The importing module can then call the imported
module as if it was in the same VM instance. In reality they execute in separate KVM vcpu instance.

## IFC (Inter Function Call)

Different modes possibles, based on low level directives.

- a _function_ has an array of input buffers (by default 1)
- a _function_ output is done in _one_ output buffer

The orchestrator knows what to do with the input and output buffers when it executes the function.
There is an API to control the execution plan done by the orchestrator.

The most basic execution plan is this one : the _caller_ gives one _input_ buffer, the _callee_ executes and the _output_ buffer is given back to the _caller_.

One more advanced execution plan is the _map-reduce_ execution plan model.

Both modes should operate through the virtual file api.

## Hooks and customization

The platforms should be very open and ease the creation of a community ecosystem of plugins and tools.

Hooks and customization allow to customize web requests processing, filter, implement ACLs and security and so on...

## Samples

The `samples` directory contains littles samples compiled from different languages, using wasi or not.

### Dashboard

This sample is an application written in javascript that uses the Guest API to provide a Web page listing all
the resources and statistics actually available from the _my own cluster_ platform.

Have the server running and run :

```bash
cd samples/dashboard
make
```

This will upload the required files and plug them to `/dashboard/` url.

You can then visit https://localhost:8443/dashboard/ in your browser.



### Watchdog

This sample is written in Rust. It is a simple Web application that uses the platform persistence utility (a key-value storage)
to store services health checks timestamps.

Have the server running and run :

```bash
cd sample/watchdog
make
```

This will upload the required files and plug them to `/watchdog/` url.

You can then visit https://localhost:8443/watchdog/ in your browser.

At the beginning no service is declared, they are added as they call their _heart-beat_ url.

You can simulate a heart beat by opening this url in another tab of your brower : https://localhost:8443/watchdog/update/SERVICE_NAME (with any _SERVICE_NAME_).

Then refresh the web page to see the updated status.



### WASI Rust demo

This is a demo of a POSIX program written in Rust running in my own cluster with the WASI ad-hoc implementation.

```bash
cd samples/wasi-rust-demo
make
```

### OpenGL compute shader demo

This is a demo using the OpenGL GLSL compute shader language. The shader simply squares float elements of an input array.

This will only work woth a real GPU (tested with NVIDIA GTX960M and GTX1060).

```bash
cd samples/gpu-compute
make
```

### Other demos

Other demos are available in the `samples/other` directory.

Here the list :

- C : `api-demo-c` (without WASI), `http-request` (with WASI), `uppercase` (with WASI)
- WebAssembly (native) : `wasi-write` (with WASI)
- Go : `hello-go` (with WASI) can be compiled with `tinygo` and normal `go` compiler.

```bash
cd samples/other
make
```


More samples are coming...


## Cleaning database

Run `make clean-db` to remove all persistence files. Your state will be lost.

## Done

- provide way to publish files with persistence
- allow the function to call another function
- run code: golang (tinygo)
- run code: rust
- run code: C/C++ (wasicc)
- run code: javascript (bare duktape runtime)
- define guest/host api over wasi/libc
- provide persistent key-value storage
- allow function to read/write in key-value storage
- Provide GPU access (through WebGPU or other)
- finish a mandelbrot sample with opengl
- integrate esgl3 for broader compatibility

## TODO

- auto reload
- update process
- multi input buffers
- sequencer/orchestrator, cost estimation and exec plan.
- ACME protocol for TLS certs auto upgrade with Let'sEncrypt
- use host name from http request in routing
- integrate client machine into the exec model (a client with webgl or js engine or wasm engine could execute some processing for local speed-up with the same processing being launched also on server later for verification. With a consensus algo maybe blockchain like to prevent fraud).
- upgrading, hot-reloading
- demo of ACL filter
- begin distribution of nodes
- sqlite engine
- include project templates in the CLI : rust, c, js, ...
- isolate wasm3 execution in KVM
- run code: compile GNU core-utils and run them to see improve WASI POSIX compatibility issues (https://github.com/coreutils/coreutils)
- run code: C#
- run code: Lua
- run code: AssemblyScript
- run code: Python
- run code: Ruby
- run code: Kotlin/Native
- run code: PHP
- allow for easy backup and disaster recovery, by default start on previous state
- implement cron like scheduling api
- think about distributing, replicating and scaling
- export and import to and from a tar.gz file (even if leveldb files should be enough...)
- SSO, User Federation, and security implemented as custom hooks

## TO SORT Usefult links

KVM, ELF, ...

- https://reverseengineering.stackexchange.com/questions/21119/how-do-tools-like-objdump-find-names-of-functions-and-their-start-address-in-elf
- https://wiki.osdev.org/Global_Descriptor_Table
- https://wiki.osdev.org/Paging
- http://ref.x86asm.net/coder32.html#xC7
- https://github.com/cirosantilli/x86-bare-metal-examples/blob/5c672f73884a487414b3e21bd9e579c67cd77621/common.h
- https://www.cs.rutgers.edu/~pxk/416/notes/
- https://stackoverflow.com/questions/18431261/how-does-x86-paging-work
- https://en.wikipedia.org/wiki/Control_register#CR4
- https://forum.osdev.org/viewtopic.php?t=11093
- https://slideplayer.com/slide/4821952/
- https://wiki.osdev.org/Entering_Long_Mode_Directly
- https://gitlab.com/bztsrc/bootboot (bof)
- https://wiki.osdev.org/Long_Mode
- https://wiki.osdev.org/CPU_Registers_x86
- https://www.kernel.org/doc/Documentation/x86/boot.txt
- https://www.st.com/content/ccc/resource/technical/document/user_manual/group1/fb/cb/d6/71/03/25/42/a1/UserManual_GNU_Assembler/files/UserManual_GNU_Assembler.pdf/jcr:content/translations/en.UserManual_GNU_Assembler.pdf
- https://github.com/cirosantilli/linux-kernel-module-cheat#userland-assembly
- https://github.com/cirosantilli/x86-bare-metal-examples#nasm
- https://github.com/soulxu/kvmsample/blob/master/main.c
- https://www.nayuki.io/page/a-fundamental-introduction-to-x86-assembly-programming
- https://www.cs.dartmouth.edu/sergey/cs258/tiny-guide-to-x86-assembly.pdf
- http://flint.cs.yale.edu/cs421/papers/x86-asm/asm.html
- https://www.cs.virginia.edu/~evans/cs216/guides/x86.html
- https://github.com/firecracker-microvm/firecracker/blob/master/src/vm-memory/src/guest_memory.rs
- https://www.linux-kvm.org/page/Simple_shell_script_to_manage_your_virtual_machine_with_bridged_networking
- https://www.kernel.org/doc/Documentation/x86/boot.txt
- https://www.kernel.org/doc/html/latest/hwmon/index.html
- https://wiki.linuxfoundation.org/networking/bridge
- https://0xax.gitbooks.io/linux-insides/content/Booting/linux-bootstrap-4.html
- https://0xax.gitbooks.io/linux-insides/content/
- https://github.com/0xAX/linux-insides/blob/master/SUMMARY.md
- https://software.intel.com/en-us/articles/intel-sdm (for the hardcore practitioner)
- https://software.intel.com/sites/default/files/managed/7c/f1/253668-sdm-vol-3a.pdf (paging described section 4.5, table 4.15)

Rust

- https://doc.rust-lang.org/nomicon/ffi.html
- https://rustwasm.github.io/docs/book/reference/js-ffi.html

WebAssembly

- https://github.com/wasienv/wasienv
- https://medium.com/wasmer/wasienv-wasi-development-workflow-for-humans-1811d9a50345
- https://github.com/NuxiNL/cloudlibc
- https://webassembly.org/docs/dynamic-linking/
- https://webassembly.org/docs/c-and-c++/
- https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-tutorial.md
- https://github.com/CraneStation/wasi-libc/blob/24792713d7e31cf593d7e19b943ef0c3aa26ef63/libc-top-half/musl/src/stdio/__stdio_write.c
- https://github.com/CraneStation/wasi-libc/blob/446cb3f1aa21f9b1a1eab372f82d65d19003e924/libc-bottom-half/cloudlibc/src/libc/sys/uio/writev.c
- https://developer.mozilla.org/fr/docs/WebAssembly/Understanding_the_text_format
- https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-api.md#args_sizes_get
- https://github.com/CraneStation/wasi-libc/blob/410c66070a2ca1724531558048f78851cc9d43fe/libc-bottom-half/libpreopen/libpreopen.c
- https://github.com/tinygo-org/tinygo/blob/v0.11.0/targets/wasm_exec.js
- https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-intro.md#how-can-i-write-programs-that-use-wasi
- https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-documents.md
- https://github.com/wasm3/wasm3-arduino
- https://dev.to/jeikabu/wasm-to-wasi-5866
- https://rustwasm.github.io/2019/03/28/this-week-in-rust-and-wasm-015.html
- https://github.com/bytecodealliance/wasi/blob/master/src/lib_generated.rs

Duktape:

- https://wiki.duktape.org/projectsusingduktape
- https://duktape.org/
- https://wiki.duktape.org/gettingstartedprimalitytesting
- https://duktape.org/api.html
- https://wiki.duktape.org/howtobuffers2x

Golang Web assembly

- https://github.com/golang/go/wiki/WebAssembly