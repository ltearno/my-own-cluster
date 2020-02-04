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
make
```

The program should build and start. It will listen on port 8443 with https protocol.

### Testing

In another terminal, go to the repository directory and type :

```bash
cd samples

# this will upload some webassembly compiled sources to the service instance
make test-register

# this will call the uploaded code and show some results
make test-call
```

## APIS

- GAPI : guest API
- PAPI : platform API
- WAPI : web API

### Guest API

From the _guest_ to the _my-own-cluster_ host.

To support legacy POSIX applications, the host implements most of its API through WASI. There is also a direct binding for applications not using `libc`.

File system :

- `api://input` : application input payload
- `api://output` : application output payload
- `http://` and `https://` : used by the guest application to issue a request to some JSON REST Service services.

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

## TODO

Compile GNU core-utils and run them to see compatibility issues (https://github.com/coreutils/coreutils)

- cli for pushing and calling function
- make those run on those languages :
- golang (tinygo)
- rust
- C/C++ (wasicc)
- others ? lua, javascipt, assembly script...
- define guest/host api over wasi/libc
- allow the function to call another function
- provide key-value storage
- allow function to read/write in key-value storage
- allow for easy backup and disaster recovery