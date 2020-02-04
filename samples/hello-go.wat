(module
  (type (;0;) (func (param i32 i32 i32 i32) (result i32)))
  (type (;1;) (func))
  (type (;2;) (func (param i32 i32)))
  (type (;3;) (func (param i32)))
  (import "wasi_unstable" "fd_write" (func $fd_write (type 0)))
  (func $__wasm_call_ctors (type 1))
  (func $_start (type 1)
    (local i32 i32)
    memory.size
    i32.const 16
    i32.shl
    i32.const 65618
    i32.sub
    i32.const 6
    i32.shr_u
    local.set 0
    i32.const 0
    local.set 1
    block  ;; label = @1
      loop  ;; label = @2
        local.get 0
        local.get 1
        i32.eq
        br_if 1 (;@1;)
        local.get 1
        i32.const 65618
        i32.add
        i32.const 0
        i32.store8
        local.get 1
        i32.const 1
        i32.add
        local.set 1
        br 0 (;@2;)
      end
    end
    i32.const 65606
    i32.const 12
    call $runtime.printstring
    call $runtime.printnl)
  (func $runtime.printstring (type 2) (param i32 i32)
    (local i32)
    i32.const 0
    local.set 2
    block  ;; label = @1
      loop  ;; label = @2
        local.get 2
        local.get 1
        i32.ge_s
        br_if 1 (;@1;)
        local.get 0
        local.get 2
        i32.add
        i32.load8_u
        call $runtime.putchar
        local.get 2
        i32.const 1
        i32.add
        local.set 2
        br 0 (;@2;)
      end
    end)
  (func $runtime.printnl (type 1)
    i32.const 13
    call $runtime.putchar
    i32.const 10
    call $runtime.putchar)
  (func $runtime.getFuncPtr (type 1)
    (local i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 0
    global.set 0
    local.get 0
    i64.const 0
    i64.store offset=8
    call $runtime.nilPanic
    unreachable)
  (func $runtime.nilPanic (type 1)
    call $runtime.runtimePanic
    unreachable)
  (func $runtime.runtimePanic (type 1)
    i32.const 65584
    i32.const 22
    call $runtime.printstring
    i32.const 65552
    i32.const 23
    call $runtime.printstring
    call $runtime.printnl
    unreachable
    unreachable)
  (func $runtime.putchar (type 3) (param i32)
    (local i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 1
    global.set 0
    i32.const 0
    local.get 0
    i32.store8 offset=65536
    local.get 1
    i32.const 0
    i32.store offset=12
    i32.const 1
    i32.const 65544
    i32.const 1
    local.get 1
    i32.const 12
    i32.add
    call $fd_write
    drop
    local.get 1
    i32.const 16
    i32.add
    global.set 0)
  (func $resume (type 1)
    call $runtime.getFuncPtr
    unreachable)
  (table (;0;) 1 1 anyfunc)
  (memory (;0;) 16)
  (global (;0;) (mut i32) (i32.const 65536))
  (global (;1;) i32 (i32.const 65618))
  (global (;2;) i32 (i32.const 65536))
  (global (;3;) i32 (i32.const 65618))
  (global (;4;) i32 (i32.const 1024))
  (export "memory" (memory 0))
  (export "__wasm_call_ctors" (func $__wasm_call_ctors))
  (export "_start" (func $_start))
  (export "__heap_base" (global 1))
  (export "resume" (func $resume))
  (export "__dso_handle" (global 2))
  (export "__data_end" (global 3))
  (export "__global_base" (global 4))
  (data (;0;) (i32.const 65536) "\00")
  (data (;1;) (i32.const 65544) "\00\00\01\00\01\00\00\00")
  (data (;2;) (i32.const 65552) "nil pointer dereference\00\00\00\00\00\00\00\00\00panic: runtime error: Hello world!"))
